package alertmanager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	html_tmpl "html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	text_tmpl "text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/notify/email"
	"github.com/prometheus/alertmanager/template"
	am_types "github.com/prometheus/alertmanager/types"
	prom_model "github.com/prometheus/common/model"
	"go.opentelemetry.io/otel/trace"

	"github.com/pgillich/micro-server/pkg/logger"
	"github.com/pgillich/micro-server/pkg/middleware"
	mw_client "github.com/pgillich/micro-server/pkg/middleware/client"
	mw_inner "github.com/pgillich/micro-server/pkg/middleware/inner"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"
	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/notifyer"
)

type NotifyStat struct {
	Firing   int
	Resolved int
}

var (
	ErrUnableToPrepareNotifier = errors.New("unable to prepare notifier")
	ErrAlertMismatchResolved   = errors.New("alert mismatch resolved")
)

func initNotifier(ctx context.Context, serverConfig *configs.ServerConfig, testConfig *configs.TestConfig, tr trace.Tracer) (*Notify, error) {
	_, log := logger.FromContext(ctx)
	hostname, _ := os.Hostname() //nolint:errcheck // not important
	var err error

	notify := &Notify{
		testConfig: testConfig,
		config:     serverConfig.Notifyer,
		tr:         tr,
	}
	notify.lastAlerts.Store(&map[string]api.GettableAlert{})

	alertmanagerUrl := notify.config.AlertmanagerUrl
	if alertmanagerUrl == "" {
		alertmanagerUrl, err = url.JoinPath("http://"+serverConfig.GetListenAddr(), configs.ServiceNameAlertmanager, "/api/v2")
		if err != nil {
			return nil, logger.Wrap(ErrUnableToPrepareNotifier, err)
		}
	}
	httpClient := mw_client.NewHttpClient(hostname, configs.ServiceNameNotifyer, TargetServiceName,
		buildinfo.BuildInfo, testConfig, log, slog.LevelInfo, slog.LevelInfo)

	notify.alertClient, err = api.NewClientWithResponses(
		alertmanagerUrl,
		api.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, logger.Wrap(ErrUnableToPrepareNotifier, err)
	}

	goKitLog := &GoKitAdapter{
		Ctx:      ctx,
		Logger:   log,
		LogLevel: slog.LevelInfo,
		Message:  "Notifyer",
	}

	notify.template, err = template.New(registerSprig)
	if err != nil {
		return nil, logger.Wrap(ErrUnableToPrepareNotifier, err)
	}
	notify.template.ExternalURL, err = url.ParseRequestURI(notify.config.ExternalURL)
	if err != nil {
		return nil, logger.Wrap(ErrUnableToPrepareNotifier, err)
	}
	err = templateFromContent(notify.template, notify.config.Templates)
	if err != nil {
		return nil, logger.Wrap(ErrUnableToPrepareNotifier, err)
	}

	for _, receiver := range notify.config.Receivers {
		if receiver.Name == notify.config.Route.Receiver {
			if receiver.EmailConfigs == nil {
				return nil, logger.Wrap(ErrUnableToPrepareNotifier, errors.New("only email receivers are supported"))
			}
			if len(receiver.EmailConfigs) != 1 {
				return nil, logger.Wrap(ErrUnableToPrepareNotifier, errors.New("only one email receiver is supported"))
			}
			notify.notifier = email.New(receiver.EmailConfigs[0], notify.template, goKitLog)
			notify.resolvedNotifier = email.New(receiver.EmailConfigs[0], notify.template, goKitLog)

			notify.receiver = &receiver

			return notify, nil
		}
	}

	return nil, logger.Wrap(ErrUnableToPrepareNotifier, errors.New("receivers not found"))
}

func templateFromContent(t *template.Template, tmpls []string) error {
	for _, tmpl := range tmpls {
		if err := t.Parse(strings.NewReader(tmpl)); err != nil {
			return err
		}
	}

	return nil
}

func (n *Notify) run(ctx context.Context, shutdown chan struct{}) {
	_, log := logger.FromContext(ctx, "goroutine", "EvalNotif")
	started := make(chan struct{})
	jobNum := 1
	go func() {
		log.Info("START_EVAL_NOTIF", "pollPeriodSec", n.config.PollPeriodSec)
		close(started)
		ticker := time.NewTicker(time.Duration(n.config.PollPeriodSec) * time.Second)
		if n.testConfig.NotifierStartEvalImmediately {
			n.callEvalNotif(ctx, jobNum)
		}
		for {
			select {
			case <-shutdown:
				log.Info("Shutdown")
				return
			case <-ctx.Done():
				log.Info("ctx.Done")
				return
			case <-ticker.C:
				n.callEvalNotif(ctx, jobNum)
			}
			jobNum++
		}
	}()
	<-started
}

func (s *Notify) callEvalNotif(ctx context.Context, jobNum int) {
	_, log := logger.FromContext(ctx)
	meter := middleware.GetMeter(buildinfo.BuildInfo, log)
	jobType := "backend"
	jobName := "eval_notif"
	jobID := fmt.Sprintf("#%d", jobNum)

	notifyStat, err := mw_inner.InternalMiddlewareChainTyped[NotifyStat](
		mw_inner.TryCatch(),
		mw_inner.Span(s.tr, jobID),
		mw_inner.Logger(map[string]string{
			logger.KeyService: "notify-job",
			"job_type":        jobType,
			"job_name":        jobName,
			"job_id":          jobID,
		}, slog.LevelInfo, slog.LevelDebug),
		mw_inner.Metrics(ctx, meter, "eval_notifications", "Eval Notofications", map[string]string{
			logger.KeyService: "notify-job",
			"job_type":        jobType,
			"job_name":        jobName,
		}, middleware.FirstErr),
		mw_inner.TryCatch(),
	)(func(ctx context.Context) (interface{}, error) {
		return s.evalNotif(ctx)
	})(ctx)

	reportLevel := slog.LevelInfo
	if err != nil {
		reportLevel = slog.LevelError
	}
	log.Log(ctx, reportLevel, "EVAL_NOTIF", "firing", notifyStat.Firing, "resolved", notifyStat.Resolved, logger.KeyError, err)
}

func (s *Notify) evalNotif(ctx context.Context) (NotifyStat, error) {
	_, log := logger.FromContext(ctx)
	notifyStat := NotifyStat{}
	reportAlerts := []*am_types.Alert{}
	resolvedAlerts := []*am_types.Alert{}
	newAlerts := map[string]api.GettableAlert{}
	lastAlerts := *s.lastAlerts.Load()

	alerts, err := s.getAlerts(ctx)
	if err != nil {
		return notifyStat, err
	}

	for _, alert := range *alerts {
		if _, has := lastAlerts[alert.Fingerprint]; has { // existing
			promAlert := ApiAlertToPromAlert(alert)
			if alert.Status.State != lastAlerts[alert.Fingerprint].Status.State { // updated
				if promAlert.Resolved() { // resolved (?)
					resolvedAlerts = append(resolvedAlerts, promAlert)
				} else { // firing (or pending?)
					reportAlerts = append(reportAlerts, promAlert)
				}
			}
		} else { // new
			promAlert := ApiAlertToPromAlert(alert)
			if promAlert.Resolved() { // resolved (?)
				resolvedAlerts = append(resolvedAlerts, promAlert)
			} else { // firing (or pending?)
				reportAlerts = append(reportAlerts, promAlert)
			}
		}
		newAlerts[alert.Fingerprint] = alert
	}

	for _, alert := range lastAlerts {
		if _, has := newAlerts[alert.Fingerprint]; !has { // removed, should be resolved
			promAlert := ApiAlertToPromAlert(alert)
			if promAlert.Resolved() { // resolved
				resolvedAlerts = append(resolvedAlerts, promAlert)
			} else {
				log.Warn("ERR_ALERT_MISMATCH", logger.KeyError, ErrAlertMismatchResolved, "alert", promAlert.String(), "endsAt", promAlert.EndsAt)
				// notify as resolved: patch endsAt
				promAlert.EndsAt = time.Now()
			}
		}
	}
	notifyStat = NotifyStat{
		Firing:   len(reportAlerts),
		Resolved: len(resolvedAlerts),
	}

	ctx = notify.WithReceiverName(ctx, "ReceiverName")
	/*
		groupLabels := prom_model.LabelSet{}
		for _, groupLabel := range s.notifyerConfig.Route.GroupByStr {
			groupLabels.
		}
		ctx = notify.WithGroupLabels(ctx, prom_model.LabelSet{DefaultGroupLabel: DefaultGroupLabelValue})
	*/
	if len(reportAlerts) > 0 {
		hasErr, err := s.notifier.Notify(ctx, reportAlerts...)
		if err != nil {
			return notifyStat, err
		}
		_ = hasErr
	}
	if len(resolvedAlerts) > 0 {
		hasErr, err := s.resolvedNotifier.Notify(ctx, resolvedAlerts...)
		if err != nil {
			return notifyStat, err
		}
		_ = hasErr
	}

	s.lastAlerts.Store(&newAlerts)

	return notifyStat, nil
}

func ApiAlertToPromAlert(alert api.GettableAlert) *am_types.Alert {
	generatorURL := ""
	if alert.GeneratorURL != nil {
		generatorURL = *alert.GeneratorURL
	}
	labels := prom_model.LabelSet{}
	for k, v := range alert.Labels {
		labels[prom_model.LabelName(k)] = prom_model.LabelValue(v)
	}
	annotations := prom_model.LabelSet{}
	for k, v := range alert.Annotations {
		annotations[prom_model.LabelName(k)] = prom_model.LabelValue(v)
	}

	return &am_types.Alert{
		Alert: prom_model.Alert{
			Labels:       labels,
			Annotations:  annotations,
			StartsAt:     alert.StartsAt,
			EndsAt:       alert.EndsAt,
			GeneratorURL: generatorURL,
		},
		UpdatedAt: alert.UpdatedAt,
		Timeout:   false,
	}
}

func (s *Notify) getAlerts(ctx context.Context) (*api.GettableAlerts, error) {
	alertsResp, err := s.alertClient.GetAlertsWithResponse(
		ctx, &api.GetAlertsParams{},
	)
	if err != nil {
		return nil, logger.Wrap(ErrAlertmanagerResponse, err)
	}
	if alertsResp.HTTPResponse.StatusCode != http.StatusOK || alertsResp.JSON200 == nil {
		return nil, logger.Wrap(ErrInvalidResponseStatus, errors.New(alertsResp.HTTPResponse.Status))
	}

	return alertsResp.JSON200, nil
}

// https://github.com/grafana/mimir/blob/e57c02c519455f5ec0156017e6998f0330a61c9c/vendor/github.com/grafana/alerting/receivers/email/email.go#L77

func registerSprig(text *text_tmpl.Template, html *html_tmpl.Template) {
	text.Funcs(text_tmpl.FuncMap(sprig.FuncMap()))

	html.Funcs(html_tmpl.FuncMap{
		"Subject":                 subjectTemplateFunc,
		"__dangerouslyInjectHTML": __dangerouslyInjectHTML,
	}).Funcs(html_tmpl.FuncMap(sprig.FuncMap()))
}

// subjectTemplateFunc sets the subject template (value) on the map represented by `.Subject.` (obj) so that it can be compiled and executed later.
// In addition, it executes and returns the subject template using the data represented in `.TemplateData` (data).
// This results in the template being replaced by the subject string.
//
// Copied from https://github.com/grafana/alerting/blob/main/receivers/email_sender.go
// Needed to https://github.com/grafana/alerting/tree/main/receivers/templates
func subjectTemplateFunc(obj map[string]any, data map[string]any, value string) string {
	obj["value"] = value

	titleTmpl, err := html_tmpl.New("title").Parse(value)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = titleTmpl.ExecuteTemplate(&buf, "title", data)
	if err != nil {
		return ""
	}

	subj := buf.String()
	// Since we have already executed the template, save it to subject data so we don't have to do it again later on
	obj["executed_template"] = subj
	return subj
}

// __dangerouslyInjectHTML allows marking areas of am email template as HTML safe, this will _not_ sanitize the string and will allow HTML snippets to be rendered verbatim.
// Use with absolute care as this _could_ allow for XSS attacks when used in an insecure context.
//
// It's safe to ignore gosec warning G203 when calling this function in an HTML template because we assume anyone who has write access
// to the email templates folder is an administrator.
//
// Copied from https://github.com/grafana/alerting/blob/main/receivers/email_sender.go
// Needed to https://github.com/grafana/alerting/tree/main/receivers/templates
//
// nolint:gosec,revive
func __dangerouslyInjectHTML(s string) html_tmpl.HTML {
	return html_tmpl.HTML(s)
}
