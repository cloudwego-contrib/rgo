import * as path from "path";
import * as os from "node:os";
import * as fs from "node:fs";
import * as util from "node:util";
import * as vscode from "vscode";

export interface downloadGoBin {
  installUrl: string;
  progressTitle: string;
  statusMessage: string;
}

export async function downloadRgoBin({
  installUrl,
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
      const command = `go install ${installUrl}@latest`;
      terminal.sendText(command);

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


export async function isGoCommandInstall(installUrl: string): Promise<boolean> {
  const gopath = process.env.GOPATH;
  if (!gopath) {
    vscode.window.showErrorMessage("GOPATH is not set in environment variables");
    return false;
  }

  const pasts = installUrl.split("/");
  const binName = pasts[pasts.length - 1];
  const binPath = path.join(
    gopath,
    "bin",
    os.platform() === "win32" ? `${binName}.exe` : `${binName}`
  );

  try {
    await fs.promises.access(binPath, fs.constants.F_OK);
    return true; // File exists
  } catch (err) {
    return false; // File does not exist or other error occurred
  }
}
