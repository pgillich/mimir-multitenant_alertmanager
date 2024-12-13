package test

import (
	"context"
	"log/slog"
	"net/url"
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/suite"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/alertmanager"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/logger"
	mw_client "github.com/pgillich/mimir-multitenant_alertmanager/internal/middleware/client"
	mw_client_model "github.com/pgillich/mimir-multitenant_alertmanager/internal/middleware/client/model"
	pkg_api "github.com/pgillich/mimir-multitenant_alertmanager/pkg/alertmanager/api"
	pkg_utils "github.com/pgillich/mimir-multitenant_alertmanager/pkg/utils"
	// "github.com/pgillich/mimir-multitenant_alertmanager/internal/tracing"
)

type E2ETestSuite struct {
	suite.Suite
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) TestAlerts() {
	log := logger.GetLogger(buildinfo.GetAppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)
	serverConfig := configs.ServerConfig{
		Alerts: configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
	}
	testConfig := configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader("X-Scope-OrgID"),
		},
	}

	server := runTestServerCmd(s.T(), "services", serverConfig, testConfig, []string{"alertmanager"}, log)
	defer server.cancel()

	testRootUrl, err := url.JoinPath(server.testServer.URL, "/alertmanager/api/v2")
	s.NoError(err, "testRootUrl")
	mimirClient, err := pkg_api.NewClientWithResponses(
		testRootUrl,
		pkg_api.WithHTTPClient(pkg_utils.NewHttpClient()),
	)
	s.NoError(err, "pkg_api.NewClientWithResponses")

	clientCtx := logger.NewContext(context.Background(), log)
	clientResp, err := mimirClient.GetAlertsWithResponse(clientCtx, &pkg_api.GetAlertsParams{})
	s.NoError(err, "GetAlertsWithResponse")

	bodyYaml, err := yaml.Marshal(clientResp.JSON200)
	s.NoError(err, "clientResp.JSON200")

	bodyStr := string(bodyYaml)
	s.T().Logf("Client Resp\n%s", bodyStr)

	//time.Sleep(1000 * time.Second)
}

func (s *E2ETestSuite) TestAlertGroups() {
	log := logger.GetLogger(buildinfo.GetAppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)
	serverConfig := configs.ServerConfig{
		Alerts: configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
	}
	testConfig := configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader("X-Scope-OrgID"),
		},
	}

	server := runTestServerCmd(s.T(), "services", serverConfig, testConfig, []string{"alertmanager"}, log)
	defer server.cancel()

	testRootUrl, err := url.JoinPath(server.testServer.URL, "/alertmanager/api/v2")
	s.NoError(err, "testRootUrl")
	mimirClient, err := pkg_api.NewClientWithResponses(
		testRootUrl,
		pkg_api.WithHTTPClient(pkg_utils.NewHttpClient()),
	)
	s.NoError(err, "pkg_api.NewClientWithResponses")

	clientCtx := logger.NewContext(context.Background(), log)
	clientResp, err := mimirClient.GetAlertGroupsWithResponse(clientCtx, &pkg_api.GetAlertGroupsParams{})
	s.NoError(err, "GetAlertGroupsWithResponse")

	bodyYaml, err := yaml.Marshal(clientResp.JSON200)
	s.NoError(err, "clientResp.JSON200")

	bodyStr := string(bodyYaml)
	s.T().Logf("Client Resp\n%s", bodyStr)

	//time.Sleep(1000 * time.Second)
}

func (s *E2ETestSuite) TestSilences() {
	log := logger.GetLogger(buildinfo.GetAppName(), slog.LevelDebug).With(logger.KeyTestCase, s.T().Name())
	//tracing.SetErrorHandlerLogger(log)
	serverConfig := configs.ServerConfig{
		Alerts: configs.AlertsConfig{
			AlertmanagerUrl: "http://localhost:8085/alertmanager/api/v2",
			Tenants:         []string{"devops", "app-development"},
			TenantLabel:     "tenant",
		},
	}
	testConfig := configs.TestConfig{
		CaptureTransportMode: mw_client_model.CaptureTransportModeFake,
		CaptureDir:           "../testdata/capture",
		CaptureMatchers: []mw_client_model.CaptureMatcher{
			mw_client.CaptureEqualRequestURLAndHeader("X-Scope-OrgID"),
		},
	}

	server := runTestServerCmd(s.T(), "services", serverConfig, testConfig, []string{"alertmanager"}, log)
	defer server.cancel()

	testRootUrl, err := url.JoinPath(server.testServer.URL, "/alertmanager/api/v2")
	s.NoError(err, "testRootUrl")
	mimirClient, err := pkg_api.NewClientWithResponses(
		testRootUrl,
		pkg_api.WithHTTPClient(pkg_utils.NewHttpClient()),
	)
	s.NoError(err, "pkg_api.NewClientWithResponses")

	clientCtx := logger.NewContext(context.Background(), log)
	clientResp, err := mimirClient.GetSilencesWithResponse(clientCtx, &pkg_api.GetSilencesParams{})
	s.NoError(err, "GetSilencesWithResponse")

	bodyYaml, err := yaml.Marshal(clientResp.JSON200)
	s.NoError(err, "clientResp.JSON200")

	bodyStr := string(bodyYaml)
	s.T().Logf("Client Resp\n%s", bodyStr)

	//time.Sleep(1000 * time.Second)
}
