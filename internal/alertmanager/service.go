package alertmanager

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	srv_configs "github.com/pgillich/micro-server/pkg/configs"
	"github.com/pgillich/micro-server/pkg/logger"
	mw_client "github.com/pgillich/micro-server/pkg/middleware/client"
	"github.com/pgillich/micro-server/pkg/model"
	"github.com/pgillich/micro-server/pkg/server"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"
	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/alertmanager"
)

const (
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
	return configs.ServiceNameAlertmanager
}

func (s *HttpService) Prepare(ctx context.Context, serverConfig srv_configs.ServerConfiger, testConfig srv_configs.TestConfiger,
	httpRouter chi.Router, tr trace.Tracer,
) error {
	_, log := logger.FromContext(ctx)
	hostname, _ := os.Hostname() //nolint:errcheck // not important

	var is bool
	s.serverConfig, is = serverConfig.(*configs.ServerConfig)
	if !is {
		return srv_configs.ErrFatalServerConfig
	}
	s.apiServer = &ApiServer{service: s}
	s.testConfig, is = testConfig.(*configs.TestConfig)
	if !is {
		return srv_configs.ErrFatalServerConfig
	}
	httpClient := mw_client.NewHttpClient(hostname, configs.ServiceNameAlertmanager, TargetServiceName,
		buildinfo.BuildInfo, s.testConfig, log, slog.LevelInfo, slog.LevelInfo)

	var err error
	s.mimirClient, err = api.NewClientWithResponses(
		s.serverConfig.Alerts.AlertmanagerUrl,
		api.WithHTTPClient(httpClient),
	)
	if err != nil {
		return logger.Wrap(ErrUnableToPrepareService, err)
	}

	api.HandlerWithOptions(s.apiServer, api.ChiServerOptions{
		BaseURL:    path.Join("/", configs.ServiceNameAlertmanager, "/api/v2"),
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
