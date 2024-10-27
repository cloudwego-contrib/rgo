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

package consts

const (
	LSPLogPathEnv = "RGO_LSP_LOG_PATH"
)

const (
	RGOLsp = "rgo_lsp"
)

const (
	MethodRGORestartLSP      = "custom/rgo/restart_language_server"
	MethodRGOWindowShowInfo  = "custom/rgo/window_show_info"
	MethodRGOWindowShowWarn  = "custom/rgo/window_show_warn"
	MethodRGOWindowShowError = "custom/rgo/window_show_error"
	MethodRGOProgress        = "custom/rgo/progress"
)

const (
	RGOProgressIDL = "rgo_progress_idl"
	RGOProgressSrc = "rgo_progress_src"

	RGOProgressStart = "start"
	RGOProgressStop  = "stop"
)

const (
	RGOProgressIDLNotification = "RGO fetching idl repos..."
	RGOProgressSrcNotification = "RGO generating src code..."
	RGOStartSuccessfully       = `{"message": "RGO started successfully"}`
)
