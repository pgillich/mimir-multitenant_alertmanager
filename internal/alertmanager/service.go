package alertmanager

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path"

	"github.com/go-chi/chi/v5"

	pkg_configs "github.com/pgillich/micro-server/pkg/configs"
	"github.com/pgillich/micro-server/pkg/logger"
	"github.com/pgillich/micro-server/pkg/middleware"
	mw_client "github.com/pgillich/micro-server/pkg/middleware/client"
	"github.com/pgillich/micro-server/pkg/model"
	"github.com/pgillich/micro-server/pkg/server"
	"github.com/pgillich/micro-server/pkg/tracing"
	pkg_utils "github.com/pgillich/micro-server/pkg/utils"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"
	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/alertmanager"
)

const (
	ServiceName       = "multitenant-alertmanager"
	TargetServiceName = "mimir_alertmanager"
)

var (
	ErrUnableToPrepareService = errors.New("unable to prepare service")
)

type HttpService struct {
	serverConfig *configs.ServerConfig
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

func (s *HttpService) Prepare(ctx context.Context, serverConfig pkg_configs.ServerConfiger, testConfig pkg_configs.TestConfiger,
	httpRouter chi.Router,
) error {
	_, log := logger.FromContext(ctx)
	hostname, _ := os.Hostname() //nolint:errcheck // not important

	var is bool
	s.serverConfig, is = serverConfig.(*configs.ServerConfig)
	if !is {
		return pkg_configs.ErrFatalServerConfig
	}
	s.apiServer = &ApiServer{service: s}
	s.testConfig, is = testConfig.(*configs.TestConfig)
	if !is {
		return pkg_configs.ErrFatalServerConfig
	}
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
		s.testConfig.CaptureTransportMode, s.testConfig.CaptureDir, s.testConfig.CaptureMatchers,
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
