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

import com.intellij.openapi.progress.ProgressIndicator;
import com.intellij.openapi.progress.Task;
import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;

import java.util.HashMap;
import java.util.Map;

public class RGOProgressManager {
    private final Map<String, Task.Backgroundable> activeTasks = new HashMap<>();
    private Project project;

    public void ProgressManager(Project project) {
        this.project = project;
    }

    public void startProgress(String id, String message) {
        if (activeTasks.containsKey(id)) {
            return;
        }

        Task.Backgroundable task = new Task.Backgroundable(project, message) {
            @Override
            public void run(@NotNull ProgressIndicator indicator) {
                indicator.setIndeterminate(true);
                while (activeTasks.containsKey(id)) {
                    try {
                        Thread.sleep(100);
                    } catch (InterruptedException e) {
                        break;
                    }
                }
            }
        };

        activeTasks.put(id, task);
        task.queue();
    }

    public void stopProgress(String id) {
        if (activeTasks.containsKey(id)) {
            activeTasks.remove(id); // 移除进度条
        }
    }
}
