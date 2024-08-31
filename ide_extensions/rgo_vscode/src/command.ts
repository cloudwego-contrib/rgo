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
