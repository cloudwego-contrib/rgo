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

import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node";

import {
  downloadRgoBin,
  isGoCommandInstall
} from "./download";
import { registerCommands } from "./command";

let client: LanguageClient;
let rgoConfig: vscode.WorkspaceConfiguration = null;

export async function activate(context: vscode.ExtensionContext) {
  registerCommands(context)

  let uri;
  if (vscode.window.activeTextEditor) {
    uri = vscode.window.activeTextEditor.document.uri;
  } else {
    uri = null;
  }
  rgoConfig = vscode.workspace.getConfiguration("rgo", uri);

  if (!rgoConfig.get<boolean>("useLanguageServer")) return;

  // Download rgo language server
  // if (!await isGoCommandInstall(rgoConfig.get('languageServerInstall'))) {
  //   try {
  //     await downloadRgoBin({
  //       installCommand: rgoConfig.get('languageServerInstall'),
  //       progressTitle: "Install Rgo Server",
  //       statusMessage: "Installing rgo language server...",
  //     });
  //   } catch (error) {
  //     vscode.window.showErrorMessage(
  //       "Error downloading rgo language server: " + error.message
  //     );
  //   }
  // }

  // if (!await isGoCommandInstall(rgoConfig.get('gopackagesdriverInstall'))) {
  //   try {
  //     await downloadRgoBin({
  //       installCommand: rgoConfig.get('gopackagesdriverInstall'),
  //       progressTitle: "Install Rgo Gopackagesdriver",
  //       statusMessage: "Installing rgo gopackagesdriver...",
  //     });
  //   } catch (error) {
  //     vscode.window.showErrorMessage(
  //       "Error downloading rgo gopackagesdriver: " + error.message
  //     );
  //   }
  // }

  await startRgoLspServer();
}

export async function startRgoLspServer() {
  const serverOptions: ServerOptions = {
    run: { command: path.join(__dirname, "../bin", "rgo_lsp_server") },
    debug: { command: path.join(__dirname, "../bin", "rgo_lsp_server") },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "go" }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher("**/.clientrc"),
    },
  };

  client = new LanguageClient(
    "rgoLanguageServer",
    "Rgo Language Server",
    serverOptions,
    clientOptions
  );

  await client.start().then(() => {
    vscode.window.showInformationMessage("Rgo Language Server started");
  });
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

