package test

import (
	"context"
	"log/slog"
	"net/url"
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/suite"

	"github.com/pgillich/micro-server/pkg/logger"
	mw_client "github.com/pgillich/micro-server/pkg/middleware/client"
	mw_client_model "github.com/pgillich/micro-server/pkg/middleware/client/model"
	srv_testutil "github.com/pgillich/micro-server/pkg/testutil"
	srv_utils "github.com/pgillich/micro-server/pkg/utils"
	srv_api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/api/alertmanager"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/alertmanager"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"
	// "github.com/pgillich/mimir-multitenant_alertmanager/internal/tracing"
)

type AlertmanagerSuite struct {
	suite.Suite
}

func TestAlertmanagerSuite(t *testing.T) {
	suite.Run(t, new(AlertmanagerSuite))
}

func (s *AlertmanagerSuite) TestAlerts() {
	log := logger.GetLogger(buildinfo.BuildInfo.AppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)
	serverConfig := &configs.ServerConfig{
		Alerts: &configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
	}
	testConfig := &configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader(configs.HttpHeaderXscopeorgid),
		},
	}

	server := srv_testutil.RunTestServerCmd(s.T(), "services",
		buildinfo.BuildInfo, serverConfig, testConfig, []string{"multitenant-alertmanager"}, log)
	defer server.Cancel()

	testRootUrl, err := url.JoinPath(server.TestServer.URL, "/multitenant-alertmanager/api/v2")
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

	//time.Sleep(1000 * time.Second)
}

func (s *AlertmanagerSuite) TestAlertGroups() {
	log := logger.GetLogger(buildinfo.BuildInfo.AppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)
	serverConfig := &configs.ServerConfig{
		Alerts: &configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
	}
	testConfig := &configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader(configs.HttpHeaderXscopeorgid),
		},
	}

	server := srv_testutil.RunTestServerCmd(s.T(), "services",
		buildinfo.BuildInfo, serverConfig, testConfig, []string{"multitenant-alertmanager"}, log)
	defer server.Cancel()

	testRootUrl, err := url.JoinPath(server.TestServer.URL, "/multitenant-alertmanager/api/v2")
	s.NoError(err, "testRootUrl")
	mimirClient, err := srv_api.NewClientWithResponses(
		testRootUrl,
		srv_api.WithHTTPClient(srv_utils.NewHttpClient()),
	)
	s.NoError(err, "srv_api.NewClientWithResponses")

	clientCtx := logger.NewContext(context.Background(), log)
	clientResp, err := mimirClient.GetAlertGroupsWithResponse(clientCtx, &srv_api.GetAlertGroupsParams{})
	s.NoError(err, "GetAlertGroupsWithResponse")

	bodyYaml, err := yaml.Marshal(clientResp.JSON200)
	s.NoError(err, "clientResp.JSON200")

	bodyStr := string(bodyYaml)
	s.T().Logf("Client Resp\n%s", bodyStr)

	//time.Sleep(1000 * time.Second)
}

func (s *AlertmanagerSuite) TestSilences() {
	log := logger.GetLogger(buildinfo.BuildInfo.AppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)
	serverConfig := &configs.ServerConfig{
		Alerts: &configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
	}
	testConfig := &configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader(configs.HttpHeaderXscopeorgid),
		},
	}

	server := srv_testutil.RunTestServerCmd(s.T(), "services",
		buildinfo.BuildInfo, serverConfig, testConfig, []string{"multitenant-alertmanager"}, log)
	defer server.Cancel()

	testRootUrl, err := url.JoinPath(server.TestServer.URL, "/alertmanager/api/v2")
	s.NoError(err, "testRootUrl")
	mimirClient, err := srv_api.NewClientWithResponses(
		testRootUrl,
		srv_api.WithHTTPClient(srv_utils.NewHttpClient()),
	)
	s.NoError(err, "srv_api.NewClientWithResponses")

	clientCtx := logger.NewContext(context.Background(), log)
	clientResp, err := mimirClient.GetSilencesWithResponse(clientCtx, &srv_api.GetSilencesParams{})
	s.NoError(err, "GetSilencesWithResponse")

	bodyYaml, err := yaml.Marshal(clientResp.JSON200)
	s.NoError(err, "clientResp.JSON200")

	bodyStr := string(bodyYaml)
	s.T().Logf("Client Resp\n%s", bodyStr)

	//time.Sleep(1000 * time.Second)
}
