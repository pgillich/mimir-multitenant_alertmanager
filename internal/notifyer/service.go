package alertmanager

import (
	"context"
	"errors"
	"path"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	am_config "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"go.opentelemetry.io/otel/trace"

	srv_configs "github.com/pgillich/micro-server/pkg/configs"
	"github.com/pgillich/micro-server/pkg/logger"
	"github.com/pgillich/micro-server/pkg/model"
	"github.com/pgillich/micro-server/pkg/server"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/notifyer"
)

const (
	TargetServiceName = "multitenant-alertmanager"
)

var (
	ErrUnableToPrepareService = errors.New("unable to prepare service")
)

type HttpService struct {
	testConfig *configs.TestConfig
	apiServer  *ApiServer
	notify     *Notify

	shutdown chan struct{}
}

type Notify struct {
	testConfig *configs.TestConfig

	config           *configs.NotifyerConfig
	alertClient      *api.ClientWithResponses
	lastAlerts       atomic.Pointer[map[string]api.GettableAlert]
	receiver         *am_config.Receiver
	notifier         notify.Notifier
	resolvedNotifier notify.Notifier
	template         *template.Template
	tr               trace.Tracer
}

func newHttpService() model.HttpServicer {
	return &HttpService{notify: &Notify{}}
}

func init() {
	server.RegisterHttpService(newHttpService)
}

func (s *HttpService) Name() string {
	return configs.ServiceNameNotifyer
}

func (s *HttpService) Prepare(ctx context.Context, serverConfiger srv_configs.ServerConfiger, testConfiger srv_configs.TestConfiger,
	httpRouter chi.Router, tr trace.Tracer,
) error {
	s.shutdown = make(chan struct{})

	serverConfig, is := serverConfiger.(*configs.ServerConfig)
	if !is {
		return srv_configs.ErrFatalServerConfig
	}
	s.apiServer = &ApiServer{service: s}
	s.testConfig, is = testConfiger.(*configs.TestConfig)
	if !is {
		return srv_configs.ErrFatalServerConfig
	}

	api.HandlerWithOptions(s.apiServer, api.ChiServerOptions{
		BaseURL:    path.Join("/", configs.ServiceNameNotifyer, "/api/v2"),
		BaseRouter: httpRouter,
	})

	var err error
	s.notify, err = initNotifier(ctx, serverConfig, s.testConfig, tr)
	if err != nil {
		return logger.Wrap(ErrUnableToPrepareService, err)
	}

	s.notify.run(ctx, s.shutdown)

	return nil
}

func (s *HttpService) Start(ctx context.Context) error {
	return nil
}

func (s *HttpService) Stop(ctx context.Context) error {
	close(s.shutdown)
	return nil
}
