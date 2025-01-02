package alertmanager

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path"

	"github.com/go-chi/chi/v5"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"
	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/alertmanager"
	"github.com/pgillich/mimir-multitenant_alertmanager/pkg/logger"
	"github.com/pgillich/mimir-multitenant_alertmanager/pkg/middleware"
	mw_client "github.com/pgillich/mimir-multitenant_alertmanager/pkg/middleware/client"
	"github.com/pgillich/mimir-multitenant_alertmanager/pkg/model"
	"github.com/pgillich/mimir-multitenant_alertmanager/pkg/server"
	"github.com/pgillich/mimir-multitenant_alertmanager/pkg/tracing"
	pkg_utils "github.com/pgillich/mimir-multitenant_alertmanager/pkg/utils"
)

const (
	ServiceName       = "multitenant-alertmanager"
	TargetServiceName = "mimir_alertmanager"
)

var (
	ErrUnableToPrepareService = errors.New("unable to prepare service")
)

type HttpService struct {
	serverConfig configs.ServerConfig
	testConfig   *configs.TestConfig
	apiServer    *ApiServer
	mimirClient  *api.ClientWithResponses
}

func newHttpService() model.HttpServicer {
	return &HttpService{}
}

func init() {
	server.RegisterHttpService(newHttpService)
}

func (s *HttpService) Name() string {
	return ServiceName
}

func (s *HttpService) Prepare(ctx context.Context, serverConfig configs.ServerConfig, testConfig *configs.TestConfig,
	httpRouter chi.Router,
) error {
	_, log := logger.FromContext(ctx)
	hostname, _ := os.Hostname() //nolint:errcheck // not important

	s.serverConfig = serverConfig
	s.apiServer = &ApiServer{service: s}
	s.testConfig = testConfig
	httpClient := mw_client.DecorateHttpClient(pkg_utils.NewHttpClient(),
		// Trace
		map[string]string{
			tracing.SpanKeyComponent: buildinfo.BuildInfo.AppName(),
			tracing.SpanKeyService:   ServiceName,
			tracing.SpanKeyInstance:  hostname,
		},
		// Metrics
		middleware.MetrHttpOut, middleware.MetrHttpOutDescr,
		map[string]string{
			middleware.MetrAttrService:       ServiceName,
			middleware.MetrAttrTargetService: TargetServiceName,
		},
		buildinfo.BuildInfo,
		// Log
		log, slog.LevelInfo, slog.LevelInfo,
		// Test
		testConfig.CaptureTransportMode, testConfig.CaptureDir, testConfig.CaptureMatchers,
	)

	var err error
	s.mimirClient, err = api.NewClientWithResponses(
		s.serverConfig.Alerts.AlertmanagerUrl,
		api.WithHTTPClient(httpClient),
	)
	if err != nil {
		return logger.Wrap(ErrUnableToPrepareService, err)
	}

	api.HandlerWithOptions(s.apiServer, api.ChiServerOptions{
		BaseURL:    path.Join("/", ServiceName, "/api/v2"),
		BaseRouter: httpRouter,
	})

	return nil
}

func (s *HttpService) Start(ctx context.Context) error {
	return nil
}

func (s *HttpService) Stop(ctx context.Context) error {
	return nil
}
