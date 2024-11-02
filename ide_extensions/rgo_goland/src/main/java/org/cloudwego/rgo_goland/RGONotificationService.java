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

import com.intellij.notification.NotificationGroupManager;
import com.intellij.notification.NotificationType;
import com.intellij.openapi.application.ApplicationManager;

import com.intellij.openapi.project.Project;
import com.intellij.platform.lsp.api.*;
import org.eclipse.lsp4j.jsonrpc.services.JsonNotification;
import org.jetbrains.annotations.Nullable;

import java.lang.reflect.Method;
import java.util.Collection;
import java.util.List;
import java.util.Map;

public class RGONotificationService extends Lsp4jClient {
    private final RGOProgressManager progressManager;
    private final Project project;

    public RGONotificationService(LspServerNotificationsHandler handler, Project project) {
        super(handler);
        this.progressManager = new RGOProgressManager();;
        this.project = project;
    }

    @JsonNotification("custom/rgo/progress")
    public void handleProgressNotification(Map<String, Object> params) {
        String id = (String) params.get("id");
        String message = (String) params.get("message");
        String type = (String) params.get("type");

        if ("start".equals(type)) {
            progressManager.startProgress(id, message);
        } else if ("stop".equals(type)) {
            progressManager.stopProgress(id);
        }
    }

//    // 实现处理 custom/rgo/restart_language_server 通知
//    @JsonNotification("custom/rgo/restart_language_server")
//    public void restartLanguageServer() {
//        LspServerManager serverManager = LspServerManager.getInstance(project);
//
//        // 获取所有与 LspServerSupportProvider 关联的 LSP 服务器
//        Collection<LspServer> lspServers = serverManager.getServersForProvider(LspServerSupportProvider.class);
//
//        for (LspServer lspServer : lspServers) {
//            try {
//                // 获取提供者类
//                Class<? extends LspServerSupportProvider> providerClass = lspServer.getProviderClass();
//
//                // 停止并重启每个 LSP 服务器
//                serverManager.stopAndRestartIfNeeded(providerClass);
//                NotificationGroupManager.getInstance()
//                        .getNotificationGroup("RGO LSP Notifications")
//                        .createNotification("RGO INFO", "All LSP servers have been restarted.", NotificationType.INFORMATION)
//                        .notify(project);
//           } catch (Exception e) {
//                NotificationGroupManager.getInstance()
//                        .getNotificationGroup("RGO LSP Notifications")
//                        .createNotification("RGO ERROR", "Failed to restart LSP servers: " + e.getMessage(), NotificationType.ERROR)
//                        .notify(project);
//            }
//        }
//    }

    // 实现处理 custom/rgo/window_show_info 通知
    @JsonNotification("custom/rgo/window_show_info")
    public void windowShowInfo(@Nullable Object params) {
        if (params instanceof Map) {
            Map<?, ?> paramMap = (Map<?, ?>) params;
            String message = (String) paramMap.get("message");

            ApplicationManager.getApplication().invokeLater(() -> {
                NotificationGroupManager.getInstance()
                        .getNotificationGroup("RGO LSP Notifications")
                        .createNotification("RGO INFO", message, NotificationType.INFORMATION)
                        .notify(project);
            });
        }
    }

    // 实现处理 custom/rgo/window_show_warn 通知
    @JsonNotification("custom/rgo/window_show_warn")
    public void windowShowWarn(@Nullable Object params) {
        // 处理警告提示的逻辑
        if (params instanceof Map) {
            Map<?, ?> paramMap = (Map<?, ?>) params;
            String message = (String) paramMap.get("message");
            ApplicationManager.getApplication().invokeLater(() -> {
                NotificationGroupManager.getInstance()
                        .getNotificationGroup("RGO LSP Notifications")
                        .createNotification("RGO WARNING", message, NotificationType.WARNING)
                        .notify(project);
            });
        }
    }

    // 实现处理 custom/rgo/window_show_error 通知
    @JsonNotification("custom/rgo/window_show_error")
    public void windowShowError(@Nullable Object params) {
        // 处理错误提示的逻辑
        if (params instanceof Map) {
            Map<?, ?> paramMap = (Map<?, ?>) params;
            String message = (String) paramMap.get("message");
            ApplicationManager.getApplication().invokeLater(() -> {
                NotificationGroupManager.getInstance()
                        .getNotificationGroup("RGO LSP Notifications")
                        .createNotification("RGO ERROR", message, NotificationType.ERROR)
                        .notify(project);
            });
        }
    }
}
