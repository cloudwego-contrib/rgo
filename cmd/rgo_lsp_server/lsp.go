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
	"io"
	"log"
	"os"

	"github.com/cloudwego-contrib/rgo/pkg/consts"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

var logPath string

func init() {
	var logger *log.Logger
	defer func() {
		logs.Init(logger)
	}()
	logPath = os.Getenv(consts.LSPLogPathEnv)
	if logPath == "" {
		logger = log.New(os.Stderr, consts.RGOLsp, 0)
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		panic(err)
	}

	multiWriter := io.MultiWriter(os.Stderr, f)
	logger = log.New(multiWriter, consts.RGOLsp, 0)
}

func RunLspServer(cancel context.CancelFunc) {
	server := lsp.NewServer(&lsp.Options{CompletionProvider: &defines.CompletionOptions{
		TriggerCharacters: &[]string{"."},
	}})

	setAllMethodsNull(server)

	server.Run()

	cancel()
}

func setAllMethodsNull(s *lsp.Server) {
	s.OnShutdown(func(ctx context.Context, req *interface{}) (err error) {
		return nil
	})

	s.OnExit(func(ctx context.Context, req *interface{}) (err error) {
		return nil
	})

	s.OnDidChangeConfiguration(func(ctx context.Context, req *defines.DidChangeConfigurationParams) (err error) {
		return nil
	})

	s.OnDidChangeWatchedFiles(func(ctx context.Context, req *defines.DidChangeWatchedFilesParams) (err error) {
		return nil
	})

	s.OnDidOpenTextDocument(func(ctx context.Context, req *defines.DidOpenTextDocumentParams) (err error) {
		return nil
	})

	s.OnDidChangeTextDocument(func(ctx context.Context, req *defines.DidChangeTextDocumentParams) (err error) {
		return nil
	})

	s.OnDidCloseTextDocument(func(ctx context.Context, req *defines.DidCloseTextDocumentParams) (err error) {
		return nil
	})

	s.OnWillSaveTextDocument(func(ctx context.Context, req *defines.WillSaveTextDocumentParams) (err error) {
		return nil
	})

	s.OnDidSaveTextDocument(func(ctx context.Context, req *defines.DidSaveTextDocumentParams) (err error) {
		return nil
	})

	s.OnExecuteCommand(func(ctx context.Context, req *defines.ExecuteCommandParams) (err error) {
		return nil
	})

	s.OnHover(func(ctx context.Context, req *defines.HoverParams) (result *defines.Hover, err error) {
		return nil, nil
	})

	s.OnCompletion(func(ctx context.Context, req *defines.CompletionParams) (result *[]defines.CompletionItem, err error) {
		return nil, nil
	})

	s.OnCompletionResolve(func(ctx context.Context, req *defines.CompletionItem) (result *defines.CompletionItem, err error) {
		return nil, nil
	})

	s.OnSignatureHelp(func(ctx context.Context, req *defines.SignatureHelpParams) (result *defines.SignatureHelp, err error) {
		return nil, nil
	})

	s.OnDeclaration(func(ctx context.Context, req *defines.DeclarationParams) (result *[]defines.LocationLink, err error) {
		return nil, nil
	})

	s.OnTypeDefinition(func(ctx context.Context, req *defines.TypeDefinitionParams) (result *[]defines.LocationLink, err error) {
		return nil, nil
	})

	s.OnImplementation(func(ctx context.Context, req *defines.ImplementationParams) (result *[]defines.LocationLink, err error) {
		return nil, nil
	})

	s.OnReferences(func(ctx context.Context, req *defines.ReferenceParams) (result *[]defines.Location, err error) {
		return nil, nil
	})

	s.OnDocumentHighlight(func(ctx context.Context, req *defines.DocumentHighlightParams) (result *[]defines.DocumentHighlight, err error) {
		return nil, nil
	})

	s.OnDocumentSymbolWithSliceDocumentSymbol(func(ctx context.Context, req *defines.DocumentSymbolParams) (result *[]defines.DocumentSymbol, err error) {
		return nil, nil
	})

	s.OnDocumentSymbolWithSliceSymbolInformation(func(ctx context.Context, req *defines.DocumentSymbolParams) (result *[]defines.SymbolInformation, err error) {
		return nil, nil
	})

	s.OnWorkspaceSymbol(func(ctx context.Context, req *defines.WorkspaceSymbolParams) (result *[]defines.SymbolInformation, err error) {
		return nil, nil
	})

	s.OnCodeActionWithSliceCommand(func(ctx context.Context, req *defines.CodeActionParams) (result *[]defines.Command, err error) {
		return nil, nil
	})

	s.OnCodeActionWithSliceCodeAction(func(ctx context.Context, req *defines.CodeActionParams) (result *[]defines.CodeAction, err error) {
		return nil, nil
	})

	s.OnCodeActionResolve(func(ctx context.Context, req *defines.CodeAction) (result *defines.CodeAction, err error) {
		return nil, nil
	})

	s.OnCodeLens(func(ctx context.Context, req *defines.CodeLensParams) (result *[]defines.CodeLens, err error) {
		return nil, nil
	})

	s.OnCodeLensResolve(func(ctx context.Context, req *defines.CodeLens) (result *defines.CodeLens, err error) {
		return nil, nil
	})

	s.OnDocumentFormatting(func(ctx context.Context, req *defines.DocumentFormattingParams) (result *[]defines.TextEdit, err error) {
		return nil, nil
	})

	s.OnDocumentRangeFormatting(func(ctx context.Context, req *defines.DocumentRangeFormattingParams) (result *[]defines.TextEdit, err error) {
		return nil, nil
	})

	s.OnDocumentOnTypeFormatting(func(ctx context.Context, req *defines.DocumentOnTypeFormattingParams) (result *[]defines.TextEdit, err error) {
		return nil, nil
	})

	s.OnRenameRequest(func(ctx context.Context, req *defines.RenameParams) (result *defines.WorkspaceEdit, err error) {
		return nil, nil
	})

	s.OnPrepareRename(func(ctx context.Context, req *defines.PrepareRenameParams) (result *defines.Range, err error) {
		return nil, nil
	})

	s.OnDocumentLinks(func(ctx context.Context, req *defines.DocumentLinkParams) (result *[]defines.DocumentLink, err error) {
		return nil, nil
	})

	s.OnDocumentLinkResolve(func(ctx context.Context, req *defines.DocumentLink) (result *defines.DocumentLink, err error) {
		return nil, nil
	})

	s.OnDocumentColor(func(ctx context.Context, req *defines.DocumentColorParams) (result *[]defines.ColorInformation, err error) {
		return nil, nil
	})

	s.OnColorPresentation(func(ctx context.Context, req *defines.ColorPresentationParams) (result *[]defines.ColorPresentation, err error) {
		return nil, nil
	})

	s.OnFoldingRanges(func(ctx context.Context, req *defines.FoldingRangeParams) (result *[]defines.FoldingRange, err error) {
		return nil, nil
	})

	s.OnSelectionRanges(func(ctx context.Context, req *defines.SelectionRangeParams) (result *[]defines.SelectionRange, err error) {
		return nil, nil
	})
}
