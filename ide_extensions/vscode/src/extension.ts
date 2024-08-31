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

  if (!rgoConfig.get<boolean>("enable")) return;

  // Download rgo language server
  // if (!await isGoCommandInstall(rgoConfig.get('lsp'))) {
  //   try {
  //     await downloadRgoBin({
  //       installCommand: rgoConfig.get('lsp'),
  //       progressTitle: "Install Rgo Server",
  //       statusMessage: "Installing rgo language server...",
  //     });
  //   } catch (error) {
  //     vscode.window.showErrorMessage(
  //       "Error downloading rgo language server: " + error.message
  //     );
  //   }
  // }

  // if (!await isGoCommandInstall(rgoConfig.get('gopackagesdriver'))) {
  //   try {
  //     await downloadRgoBin({
  //       installCommand: rgoConfig.get('gopackagesdriver'),
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
    run: { command: path.join(__dirname, "../bin", "go-lsp") },
    debug: { command: path.join(__dirname, "../bin", "go-lsp") },
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

