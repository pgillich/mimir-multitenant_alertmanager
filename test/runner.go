package test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pgillich/mimir-multitenant_alertmanager/cmd"
	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/alertmanager"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/logger"
	// "github.com/pgillich/mimir-multitenant_alertmanager/internal/tracing"
)

type TestServer struct {
	testServer *httptest.Server
	addr       string
	ctx        context.Context //nolint:containedctx // test
	cancel     context.CancelFunc
}

func runTestServerCmd(t *testing.T, serverName string, serverConfig configs.ServerConfig, testConfig configs.TestConfig, args []string, log *slog.Logger) *TestServer {
	ctx := logger.NewContext(context.Background(), log)

	server := &TestServer{
		testServer: httptest.NewUnstartedServer(nil),
	}
	server.addr = server.testServer.Listener.Addr().String()

	started := make(chan struct{})
	runner := HttpTestserverRunner(server.testServer, started)
	server.ctx, server.cancel = context.WithCancel(ctx)
	testConfig.HttpServerRunner = runner

	workDir := t.TempDir()

	serverConfigFile, err := configs.SaveServerConfig(serverConfig, workDir, "multitenant_alertmanager.yaml")
	if err != nil {
		t.Fatalf("SaveServerConfig: %v", err)
	}

	command := append([]string{
		serverName,
		"--config", serverConfigFile,
		"--listenaddr", server.addr,
	}, args...)

	go func() {
		cmd.Execute(server.ctx, command, testConfig)
	}()
	<-started
	//time.Sleep(1 * time.Second)

	return server
}

func HttpTestserverRunner(server *httptest.Server, started chan struct{}) configs.HttpServerRunner {
	return func(h http.Handler, shutdown <-chan struct{}, addr string, log *slog.Logger) {
		server.Config.Handler = h
		log.Info("TestServer start")
		server.Start()
		close(started)
		log.Info("TestServer started")
		<-shutdown
		log.Info("TestServer shutdown")
		server.Close()
	}
}
