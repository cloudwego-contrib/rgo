package org.cloudwego.rgo_goland;

import com.intellij.execution.process.KillableColoredProcessHandler;
import com.intellij.execution.process.ProcessEvent;
import com.intellij.execution.process.ProcessListener;
import com.intellij.execution.process.ProcessOutputTypes;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.platform.lang.lsWidget.LanguageServiceWidgetItem;
import com.intellij.platform.lsp.api.*;
import com.intellij.platform.lsp.api.customization.LspCommandsSupport;
import com.intellij.platform.lsp.api.lsWidget.LspServerWidgetItem;
import org.eclipse.lsp4j.Command;
import org.jetbrains.annotations.NotNull;
import com.intellij.execution.configurations.GeneralCommandLine;
import org.jetbrains.annotations.Nullable;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;


public class RgoLspServerService implements LspServerSupportProvider {
    @Override
    public void fileOpened(Project project, VirtualFile file, LspServerStarter serverStarter) {
        if ("go".equals(file.getExtension())) {
            RgoLspServerDescriptor descriptor = new RgoLspServerDescriptor(project);
            serverStarter.ensureServerStarted(descriptor);
        }
    }


    private static class RgoLspServerDescriptor extends ProjectWideLspServerDescriptor {

        private final Project project;

        public RgoLspServerDescriptor(Project project) {
            super(project, "rgo");
            this.project = project;
        }

        @Override
        public @NotNull Lsp4jClient createLsp4jClient(@NotNull LspServerNotificationsHandler handler) {
            return new RGONotificationService(handler,project);
        }

        @Override
        public boolean isSupportedFile(VirtualFile file) {
            return "go".equals(file.getExtension());
        }

        @Override
        public GeneralCommandLine createCommandLine() {
            try {
                // 从 resources 提取到临时目录
                File tempExecutable = extractResourceToTemp("bin/rgo_lsp_server");

                // 设置可执行权限
                tempExecutable.setExecutable(true);

                return new GeneralCommandLine(tempExecutable.getAbsolutePath())
                        .withWorkDirectory(this.getProject().getBasePath());
            } catch (IOException e) {
                throw new RuntimeException("Failed to extract or set up executable", e);
            }
        }

        @Override
        public LspCommandsSupport getLspCommandsSupport() {
            return new RgoCommandSupport(super.getProject());
        }

        private File extractResourceToTemp(String resourcePath) throws IOException {
            InputStream resourceStream = getClass().getClassLoader().getResourceAsStream(resourcePath);
            if (resourceStream == null) {
                throw new IOException("Resource not found: " + resourcePath);
            }

            File tempFile = File.createTempFile("rgo_lsp_server", null);
            try (FileOutputStream out = new FileOutputStream(tempFile)) {
                byte[] buffer = new byte[1024];
                int bytesRead;
                while ((bytesRead = resourceStream.read(buffer)) != -1) {
                    out.write(buffer, 0, bytesRead);
                }
            }
            return tempFile;
        }
    }
}

