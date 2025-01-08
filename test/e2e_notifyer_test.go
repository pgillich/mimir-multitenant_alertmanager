package test

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"testing"
	"time"

	smtp "github.com/emersion/go-smtp"
	yaml "github.com/goccy/go-yaml"
	am_config "github.com/prometheus/alertmanager/config"
	"github.com/stretchr/testify/suite"

	"github.com/pgillich/micro-server/pkg/logger"
	mw_client "github.com/pgillich/micro-server/pkg/middleware/client"
	mw_client_model "github.com/pgillich/micro-server/pkg/middleware/client/model"
	srv_testutil "github.com/pgillich/micro-server/pkg/testutil"
	srv_utils "github.com/pgillich/micro-server/pkg/utils"
	srv_api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/alertmanager"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"

	// "github.com/pgillich/mimir-multitenant_alertmanager/internal/tracing"

	// force to run init() functions
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/alertmanager"
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/notifyer"
)

type NotifyerSuite struct {
	suite.Suite
}

func TestNotifyerSuite(t *testing.T) {
	suite.Run(t, new(NotifyerSuite))
}

func (s *NotifyerSuite) TestNotify0() {
	log := logger.GetLogger(buildinfo.BuildInfo.AppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)

	defaultTmpl, err := os.ReadFile("../testdata/notifier/default.tmpl")
	s.NoError(err, "default.tmpl")
	emailTmpl, err := os.ReadFile("../testdata/notifier/email.tmpl")
	s.NoError(err, "email.tmpl")
	extensionTmpl, err := os.ReadFile("../testdata/notifier/extension.tmpl")
	s.NoError(err, "extension.tmpl")
	txtTmpl, err := os.ReadFile("../testdata/notifier/ng_alert_notification.txt")
	s.NoError(err, "ng_alert_notification.txt")
	emailRequireTLS := false
	serverConfig := &configs.ServerConfig{
		Alerts: &configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
		Notifyer: &configs.NotifyerConfig{
			AlertmanagerUrl: "", // Default, will be filled in Prepare
			ExternalURL:     "http://ExternalURL",
			PollPeriodSec:   600,
			Route: &am_config.Route{
				Receiver: "email",
				// GroupByStr: []string{"severity"}, // uses Dispatcher? []string{"alertname", "severity"},
			},
			Receivers: []am_config.Receiver{
				{
					Name: "email",
					EmailConfigs: []*am_config.EmailConfig{
						{
							Smarthost:    am_config.HostPort{Host: "localhost", Port: "2525"},
							To:           "testuser@localhost", // :2525
							From:         "notifier@e2e.test",
							AuthUsername: "testuser",
							AuthPassword: "testpass",
							RequireTLS:   &emailRequireTLS,
							//HTML:         "Hello, <b>world</b>!",
							Text:    string(txtTmpl),
							Headers: map[string]string{},
						},
					},
				},
			},
			Templates: []string{
				string(defaultTmpl),
				string(emailTmpl),
				string(extensionTmpl),
			},
		},
	}
	testConfig := &configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader(configs.HttpHeaderXscopeorgid),
		},
		NotifierStartEvalImmediately: true,
	}

	smtpServer := smtp.NewServer(&SmtpBackend{Logger: log})
	smtpServer.Addr = ":2525"
	smtpServer.Domain = "localhost"
	smtpServer.WriteTimeout = 600 * time.Second // 10
	smtpServer.ReadTimeout = 600 * time.Second  // 10
	smtpServer.MaxMessageBytes = 1024 * 1024
	smtpServer.MaxRecipients = 50
	smtpServer.AllowInsecureAuth = true
	smtpStarted := make(chan struct{})
	go func() {
		close(smtpStarted)
		err = smtpServer.ListenAndServe()
		s.NoError(err, "smtpServer.ListenAndServe")

		defer smtpServer.Shutdown(context.Background())
	}()
	<-smtpStarted
	time.Sleep(1 * time.Second)

	server := srv_testutil.RunTestServerCmd(s.T(), "services",
		buildinfo.BuildInfo, serverConfig, testConfig, []string{"multitenant-alertmanager", "notifyer"}, log)
	defer server.Cancel()

	testRootUrl, err := url.JoinPath(server.TestServer.URL, "/notifyer/api/v2")
	s.NoError(err, "testRootUrl")
	mimirClient, err := srv_api.NewClientWithResponses(
		testRootUrl,
		srv_api.WithHTTPClient(srv_utils.NewHttpClient()),
	)
	s.NoError(err, "srv_api.NewClientWithResponses")

	clientCtx := logger.NewContext(context.Background(), log)
	clientResp, err := mimirClient.GetAlertsWithResponse(clientCtx, &srv_api.GetAlertsParams{})
	s.NoError(err, "GetAlertsWithResponse")

	bodyYaml, err := yaml.Marshal(clientResp.JSON200)
	s.NoError(err, "clientResp.JSON200")

	bodyStr := string(bodyYaml)
	s.T().Logf("Client Resp\n%s", bodyStr)

	time.Sleep(1000 * time.Second)
}
