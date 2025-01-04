package configs

import (
	pkg_configs "github.com/pgillich/micro-server/pkg/configs"

	mw_client_model "github.com/pgillich/micro-server/pkg/middleware/client/model"
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

func (c *ServerConfig) GetListenAddr() string {
	return c.ListenAddr
}

func (c *ServerConfig) GetTracerUrl() string {
	return c.TracerUrl
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
	HttpServerRunner     pkg_configs.HttpServerRunner
}

func (c *TestConfig) GetHttpServerRunner() pkg_configs.HttpServerRunner {
	return c.HttpServerRunner
}

func (c *TestConfig) SetHttpServerRunner(httpServerRunner pkg_configs.HttpServerRunner) {
	c.HttpServerRunner = httpServerRunner
}
