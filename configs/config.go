package configs

import (
	srv_configs "github.com/pgillich/micro-server/pkg/configs"

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
	HttpServerRunner     srv_configs.HttpServerRunner
}

func (c *TestConfig) GetCaptureTransportMode() mw_client_model.CaptureTransportMode {
	return c.CaptureTransportMode
}

func (c *TestConfig) GetCaptureDir() string {
	return c.CaptureDir
}

func (c *TestConfig) GetCaptureMatchers() []mw_client_model.CaptureMatcher {
	return c.CaptureMatchers
}

func (c *TestConfig) GetHttpServerRunner() srv_configs.HttpServerRunner {
	return c.HttpServerRunner
}

func (c *TestConfig) SetHttpServerRunner(httpServerRunner srv_configs.HttpServerRunner) {
	c.HttpServerRunner = httpServerRunner
}
