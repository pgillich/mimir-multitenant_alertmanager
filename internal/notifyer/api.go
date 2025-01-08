package alertmanager

import (
	"errors"
	"maps"
	"net/http"
	"slices"

	"github.com/pgillich/micro-server/pkg/logger"

	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/notifyer"
)

type ApiServer struct {
	service *HttpService
}

var (
	ErrAlertmanagerResponse  = errors.New("alertmanager response")
	ErrInvalidResponseStatus = errors.New("invalid response status")
	ErrRenderResponse        = errors.New("unable to render response")
)

func (s *ApiServer) GetAlerts(w http.ResponseWriter, r *http.Request, params api.GetAlertsParams) {
	_, log := logger.FromContext(r.Context(), "alertmanagerUrl", s.service.notify.config.AlertmanagerUrl)
	alerts := s.service.notify.lastAlerts.Load()

	if err := api.GetAlerts200JSONResponse(slices.Collect(maps.Values(*alerts))).VisitGetAlertsResponse(w); err != nil {
		log.Error("Unable to render response", logger.KeyError, logger.Wrap(ErrRenderResponse, err))
	}
}
