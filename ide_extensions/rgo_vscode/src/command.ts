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

import * as vscode from "vscode";
import { downloadRgoBin } from "./download";
import { startRgoLspServer } from "./extension";

export function registerCommands(context: vscode.ExtensionContext) {
  const rgoConfig = vscode.workspace.getConfiguration("rgo");
  context.subscriptions.push(
    vscode.commands.registerCommand("rgo.install", async () => downloadRgoBin({
      installCommand: rgoConfig.get('url'),
      progressTitle: "Install Rgo Server",
      statusMessage: "Installing rgo language",
    }))
  );
  context.subscriptions.push(
    vscode.commands.registerCommand("rgo.restart", startRgoLspServer)
  );
  context.subscriptions.push(
    vscode.commands.registerCommand(
      "rgo.gopackagesdriver",
      async () => downloadRgoBin({
        installCommand: rgoConfig.get('gopackagesdriver'),
        progressTitle: "Install Rgo Gopackagesdriver",
        statusMessage: "Installing rgo gopackagesdriver",
      })
    )
  );
}
