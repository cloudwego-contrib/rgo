/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"

	"github.com/cloudwego-contrib/rgo/pkg/rlog"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	server := lsp.NewServer(&lsp.Options{CompletionProvider: &defines.CompletionOptions{
		TriggerCharacters: &[]string{"."},
	}})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())
				rlog.Error("Recovered from panic in RGORun", zap.Any("error", r), zap.String("stack_trace", stackTrace))
			}
		}()

		RGORun(ctx)
	}()

	// gracefully shutdown
	go func() {
		sig := <-signalChan
		rlog.Info("Received signal, shutting down...", zap.String("signal", sig.String()))

		cancel()

		time.Sleep(2 * time.Second)

		os.Exit(0)
	}()

	RunLspServer(cancel, server)
}
