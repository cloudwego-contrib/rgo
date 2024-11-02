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
