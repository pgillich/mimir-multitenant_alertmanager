package alertmanager

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	prom_model "github.com/prometheus/common/model"

	"github.com/pgillich/micro-server/pkg/logger"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/alertmanager"
)

type ApiServer struct {
	service *HttpService
}

var (
	ErrMimirResponse         = errors.New("mimir response")
	ErrInvalidResponseStatus = errors.New("invalid response status")
	ErrRenderResponse        = errors.New("unable to render response")
)

func RequestHeaderSet(headerKey, headerValue string) func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set(headerKey, headerValue)
		return nil
	}
}

func (s *ApiServer) GetAlerts(w http.ResponseWriter, r *http.Request, params api.GetAlertsParams) {
	_, log := logger.FromContext(r.Context(), "alertmanagerUrl", s.service.serverConfig.Alerts.AlertmanagerUrl)
	alerts := []api.GettableAlert{}

	for _, tenant := range s.service.serverConfig.Alerts.Tenants {
		mimirResp, err := s.service.mimirClient.GetAlertsWithResponse(
			r.Context(), &params, RequestHeaderSet(configs.HttpHeaderXscopeorgid, tenant),
		)
		if err != nil {
			err = logger.Wrap(ErrMimirResponse, err)
			log.Error("Unable to GetAlerts", logger.KeyError, err)
			if err = api.GetAlerts500JSONResponse(err.Error()).VisitGetAlertsResponse(w); err != nil {
				log.Error("Unable to render error response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
			}
			return
		}
		if mimirResp.HTTPResponse.StatusCode != http.StatusOK || mimirResp.JSON200 == nil {
			err = logger.Wrap(ErrInvalidResponseStatus, errors.New(mimirResp.HTTPResponse.Status))
			log.Error("Unable to GetAlerts", logger.KeyError, err)
			if err = api.GetAlerts500JSONResponse(err.Error()).VisitGetAlertsResponse(w); err != nil {
				log.Error("Unable to render error response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
			}
			return
		}
		for _, alert := range *mimirResp.JSON200 {
			mustFingerprint := strconv.FormatUint(prom_model.LabelsToSignature(alert.Labels), 16)
			if alert.Fingerprint != mustFingerprint {
				log.Debug("Fingerprint mismatch", "alertFingerprint", alert.Fingerprint, "mustFingerprint", mustFingerprint)
			}
			alert.Annotations[s.service.serverConfig.Alerts.TenantLabel] = tenant
			alert.Labels[s.service.serverConfig.Alerts.TenantLabel] = tenant
			alert.Fingerprint = strconv.FormatUint(prom_model.LabelsToSignature(alert.Labels), 16)
			for r := range alert.Receivers {
				alert.Receivers[r].Name = tenant + "/" + alert.Receivers[r].Name
			}
			alerts = append(alerts, alert)
		}
	}

	if err := api.GetAlerts200JSONResponse(alerts).VisitGetAlertsResponse(w); err != nil {
		log.Error("Unable to render response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
	}
}

func (s *ApiServer) GetAlertGroups(w http.ResponseWriter, r *http.Request, params api.GetAlertGroupsParams) {
	_, log := logger.FromContext(r.Context(), "alertmanagerUrl", s.service.serverConfig.Alerts.AlertmanagerUrl)
	alertGroups := []api.AlertGroup{}

	for _, tenant := range s.service.serverConfig.Alerts.Tenants {
		mimirResp, err := s.service.mimirClient.GetAlertGroupsWithResponse(
			r.Context(), &params, RequestHeaderSet(configs.HttpHeaderXscopeorgid, tenant),
		)
		if err != nil {
			err = logger.Wrap(ErrMimirResponse, err)
			log.Error("Unable to GetAlerts", logger.KeyError, err)
			if err = api.GetAlertGroups500JSONResponse(err.Error()).VisitGetAlertGroupsResponse(w); err != nil {
				log.Error("Unable to render response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
			}
			return
		}
		if mimirResp.HTTPResponse.StatusCode != http.StatusOK || mimirResp.JSON200 == nil {
			err = logger.Wrap(ErrInvalidResponseStatus, errors.New(mimirResp.HTTPResponse.Status))
			log.Error("Unable to GetAlerts", logger.KeyError, err)
			if err = api.GetAlertGroups500JSONResponse(err.Error()).VisitGetAlertGroupsResponse(w); err != nil {
				log.Error("Unable to render error response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
			}
			return
		}
		for _, alertGroup := range *mimirResp.JSON200 {
			alertGroup.Labels[s.service.serverConfig.Alerts.TenantLabel] = tenant
			for a := range alertGroup.Alerts {
				alert := &alertGroup.Alerts[a]
				mustFingerprint := strconv.FormatUint(prom_model.LabelsToSignature(alert.Labels), 16)
				if alert.Fingerprint != mustFingerprint {
					log.Debug("Fingerprint mismatch", "alertFingerprint", alert.Fingerprint, "mustFingerprint", mustFingerprint)
				}
				alert.Annotations[s.service.serverConfig.Alerts.TenantLabel] = tenant
				alert.Labels[s.service.serverConfig.Alerts.TenantLabel] = tenant
				alert.Fingerprint = strconv.FormatUint(prom_model.LabelsToSignature(alert.Labels), 16)
				for r := range alert.Receivers {
					alert.Receivers[r].Name = tenant + "/" + alert.Receivers[r].Name
				}
			}
			alertGroup.Receiver.Name = tenant + "/" + alertGroup.Receiver.Name

			alertGroups = append(alertGroups, alertGroup)
		}
	}

	if err := api.GetAlertGroups200JSONResponse(alertGroups).VisitGetAlertGroupsResponse(w); err != nil {
		log.Error("Unable to render response", logger.KeyError, err)
	}
}

func (s *ApiServer) GetSilences(w http.ResponseWriter, r *http.Request, params api.GetSilencesParams) {
	_, log := logger.FromContext(r.Context(), "alertmanagerUrl", s.service.serverConfig.Alerts.AlertmanagerUrl)
	silences := []api.GettableSilence{}

	for _, tenant := range s.service.serverConfig.Alerts.Tenants {
		mimirResp, err := s.service.mimirClient.GetSilencesWithResponse(
			r.Context(), &params, RequestHeaderSet(configs.HttpHeaderXscopeorgid, tenant),
		)
		if err != nil {
			err = logger.Wrap(ErrMimirResponse, err)
			log.Error("Unable to GetSilences", logger.KeyError, err)
			if err = api.GetSilences500JSONResponse(err.Error()).VisitGetSilencesResponse(w); err != nil {
				log.Error("Unable to render error response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
			}
			return
		}
		if mimirResp.HTTPResponse.StatusCode != http.StatusOK || mimirResp.JSON200 == nil {
			err = logger.Wrap(ErrInvalidResponseStatus, errors.New(mimirResp.HTTPResponse.Status))
			log.Error("Unable to GetSilences", logger.KeyError, err)
			if err = api.GetSilences500JSONResponse(err.Error()).VisitGetSilencesResponse(w); err != nil {
				log.Error("Unable to render error response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
			}
			return
		}
		equal := true
		for _, silence := range *mimirResp.JSON200 {
			silence.Matchers = append(silence.Matchers, api.Matcher{
				Name:    s.service.serverConfig.Alerts.TenantLabel,
				Value:   tenant,
				IsEqual: &equal,
				IsRegex: false,
			})
			silences = append(silences, silence)
		}
	}

	if err := api.GetSilences200JSONResponse(silences).VisitGetSilencesResponse(w); err != nil {
		log.Error("Unable to render response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
	}
}
