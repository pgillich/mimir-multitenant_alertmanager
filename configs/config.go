package configs

import (
	"log/slog"
	"net/http"

	mw_client_model "github.com/pgillich/mimir-multitenant_alertmanager/internal/middleware/client/model"
)

const (
	TracerVersion      = "0.1.0"
	DefaultTenantLabel = "tenant"
)

type ServerConfig struct {
	ListenAddr string
	TracerUrl  string
	Alerts     AlertsConfig
}

type AlertsConfig struct {
	AlertmanagerUrl string
	Tenants         []string
	TenantLabel     string
}

type TestConfig struct {
	CaptureTransportMode mw_client_model.CaptureTransportMode
	CaptureDir           string
	CaptureMatchers      []mw_client_model.CaptureMatcher
	HttpServerRunner     HttpServerRunner
}

type HttpServerRunner func(h http.Handler, shutdown <-chan struct{}, addr string, l *slog.Logger)
