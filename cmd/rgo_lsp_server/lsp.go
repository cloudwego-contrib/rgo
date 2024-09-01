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
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

var logPath *string

func init() {
	var logger *log.Logger
	defer func() {
		logs.Init(logger)
	}()
	logPath = flag.String("logs", "", "logs file path")
	if logPath == nil || *logPath == "" {
		logger = log.New(os.Stderr, "", 0)
		return
	}
	p := *logPath
	f, err := os.Open(p)
	if err == nil {
		logger = log.New(f, "", 0)
		return
	}
	f, err = os.Create(p)
	if err == nil {
		logger = log.New(f, "", 0)
		return
	}
	panic(fmt.Sprintf("logs init error: %v", *logPath))
}

func RunLspServer(cancel context.CancelFunc) {
	server := lsp.NewServer(&lsp.Options{CompletionProvider: &defines.CompletionOptions{
		TriggerCharacters: &[]string{"."},
	}})

	server.OnCompletion(func(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
		return nil, nil
	})

	server.Run()

	cancel()
}
