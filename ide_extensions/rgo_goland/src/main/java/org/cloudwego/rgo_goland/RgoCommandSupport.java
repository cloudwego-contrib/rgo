package org.cloudwego.rgo_goland;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.platform.lsp.api.LspServer;
import com.intellij.platform.lsp.api.customization.LspCommandsSupport;
import org.eclipse.lsp4j.Command;

public class RgoCommandSupport extends LspCommandsSupport {

    private final Project project;

    public RgoCommandSupport(Project project) {
        this.project = project;
    }

    @Override
    public void executeCommand(LspServer server, VirtualFile contextFile, Command command) {
        super.executeCommand(server, contextFile, command);
    }

}
