/*
Copyright © 2024 Peter Gillich <pgillich@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/pgillich/micro-server/pkg/cmd"
	"github.com/pgillich/micro-server/pkg/logger"

	"github.com/pgillich/mimir-multitenant_alertmanager/configs"
	"github.com/pgillich/mimir-multitenant_alertmanager/internal/buildinfo"

	// force to run init() functions
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/alertmanager"
	_ "github.com/pgillich/mimir-multitenant_alertmanager/internal/notifyer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = logger.NewContext(ctx, logger.GetLogger(buildinfo.BuildInfo.AppName(), slog.LevelDebug))
	cmd.Execute(ctx, os.Args[1:], buildinfo.BuildInfo, &configs.ServerConfig{}, &configs.TestConfig{})
	cancel()
}
