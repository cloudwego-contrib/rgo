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
import * as os from "node:os";
import * as fs from "node:fs";
import * as vscode from "vscode";

export interface downloadGoBin {
  installCommand: string;
  progressTitle: string;
  statusMessage: string;
}

export async function downloadRgoBin({
  installCommand,
  progressTitle,
  statusMessage,
}: downloadGoBin) {
  const terminal = vscode.window.createTerminal();

  await vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: progressTitle,
      cancellable: false,
    },
    async (progress) => {
      terminal.sendText(installCommand);

      // Wait for the command to complete (you can adjust the timeout as needed)
      await new Promise<void>((resolve) => {
        const interval = setInterval(() => {
          // Check if the command has completed (you can customize this condition)
          const isCommandComplete = /* Your condition here */ true;

          if (isCommandComplete) {
            clearInterval(interval);
            resolve();
          }
        }, 1000); // Check every second
      });

      // Command completed, update status message
      vscode.window.setStatusBarMessage(statusMessage, 3000);
    }
  );
}


export async function isGoCommandInstall(installCommand: string): Promise<boolean> {
  const gopath = process.env.GOPATH;
  if (!gopath) {
    vscode.window.showErrorMessage("GOPATH is not set in environment variables");
    return false;
  }

  const noVersionCommand = installCommand.split('@')[0];
  const noGoInstallCommand = noVersionCommand.replace('go install', '').trim();
  
  const binaryFile = noGoInstallCommand.substring(noGoInstallCommand.lastIndexOf('/') + 1);
  if (!binaryFile) {
    vscode.window.showErrorMessage("Invalid install command");
    return false;
  }

  const binPath = path.join(
    gopath,
    "bin",
    os.platform() === "win32" ? `${binaryFile}.exe` : binaryFile
  );

  try {
    await fs.promises.access(binPath, fs.constants.F_OK);
    return true; // File exists
  } catch (err) {
    return false; // File does not exist or other error occurred
  }
}
