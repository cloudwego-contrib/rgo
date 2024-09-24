# RGO

RGO 目前处于 MVP 阶段

# 运行步骤

目前 RGO 需要运行的组件有 ：

- rgo_lsp_server —— lsp 插件
- rgopackagesdriver —— GOPACKAGESDRIVER 组件
- rgo —— 编译期运行命令行工具

目前支持 rgo_lsp_server 的 IDE 有：

- VS-Code
- Emacs
- VIM/Neovim

VS-Code 插件因为还未发布，所以目前只能在 VS-Code 本地调试运行。

## clone RGO 仓库

```shell
git clone https://github.com/cloudwego-contrib/rgo.git

cd rgo
```

## 编译生成 gopackagesdriver

```shell
cd cmd/rgopackagesdriver

go install

cd ../..
```

## 编译安装 rgo

```shell
cd cmd/rgo && go install
cd ../..
```

## 配置 rgo_lsp_server

### VS-Code

#### 编译生成 rgo_lsp_server

```shell
cd cmd/rgo_lsp_server && go build -o rgo_lsp_server .
mv rgo_lsp_server ../../ide_extensions/rgo_vscode/bin/
cd ../..
```

#### 在 VS-Code 中调试运行插件

在 VS-Code 中打开克隆下来的 ide_extensions/rgo_vscode 项目

![doc/vscode_open.png](doc/vscode_open.png)

在 rgo_vscode 项目根目录下执行以下命令安装依赖

若从未安装过 nodejs 依赖，macos 执行

```shell
brew install node
```

linux 执行

```shell
sudo apt-get install nodejs
```

执行以下命令安装 npm 依赖

```shell
npm install
```

使用快捷键 F5 即可本地启动 vscode 插件 、或参考截图

![edit_ide.png](doc/vscode-extension.png)

#### 新建测试项目

然后会弹出搭载了 go-lsp 插件的 vscode 窗口，新建一个新的项目用于展示效果

```shell
mkdir -p ~/rgo_test
cd ~/rgo_test
```

#### 在根目录下新建配置文件 rgo_config.yaml

```yaml
idl_repos:
  - repo_name: kitex_example
    git_url: https://github.com/cloudwego/kitex-examples.git
    branch: main
    commit: 
idls:
  - idl_path: hello/hello.thrift
    repo_name: kitex_example
    service_name: a.b.c

```

#### 修改 VS-Code 配置

```shell
vim .vscode/settings.json
```

填入配置并保存

```json
{
  "go.toolsEnvVars": {
    "GOPACKAGESDRIVER": "${env:GOPATH}/bin/rgopackagesdriver"
  },
  "go.enableCodeLens": {
    "runtest": false
  },
  "gopls": {
    "formatting.gofumpt": true,
    "formatting.local": "rgo/",
    "ui.completion.usePlaceholders": false,
    "ui.semanticTokens": true,
    "ui.codelenses": {
      "gc_details": false,
      "regenerate_cgo": false,
      "generate": false,
      "test": false,
      "tidy": false,
      "upgrade_dependency": false,
      "vendor": false
    }
  },
  "go.useLanguageServer": true,
  "go.buildOnSave": "off",
  "go.lintOnSave": "off",
  "go.vetOnSave": "off"
}

```

#### vscode 插件

##### 支持功能

- [x] rgo_config.yml 配置文件解析智能提示
- [x] 自动预下载 `rgo_lsp_server` && `rgopackagesdriver`
- [x] 支持**消费**配置中的 idls, 做到无侵入的代码提示

##### 支持 vscode 配置

- `rgo.useLanguageServer`, 默认为 `true`, 用于控制是否开启 rgo language server，关闭后将不会有任何提示
- `rgo.languageServerInstall`, 默认为 `go install github.com/cloudwego-contrib/cmd/rgo_lsp_server@latest`, 用于配置
  rgo_lsp_server 的安装命令
- `rgo.gopackagesdriverInstall`, 默认为 `go install github.com/cloudwego-contrib/cmd/rgopackagesdriver@latest`, 用于配置
  rgopackagesdriver 的安装命令

##### 支持的 vscode 命令

- `RGo: Install language server`, 用于安装 rgo_lsp_server
- `RGo: Install gopackagesdriver`, 用于安装 rgopackagesdriver
- `RGo: Restart language server`, 用于重启 rgo_lsp_server

![restart_gopls.png](doc/restart_gopls.png)

#### 效果展示

![show.png](doc/show_vscode.png)



### Emacs

#### 编译安装 rgo_lsp_server

```shell
go install cmd/rgo_lsp_server
```



#### 在根目录下新建配置文件 rgo_config.yaml

新建一个项目，在 rgo_config.yaml 中添加配置，示例如下：

```yaml
idl_repos:
  - repo_name: kitex_example
    git_url: https://github.com/cloudwego/kitex-examples.git
    branch: main
    commit: 
idls:
  - idl_path: hello/hello.thrift
    repo_name: kitex_example
    service_name: a.b.c

```



#### 配置 GOPACKAGESDRIVER

调用 M-x [add-dir-local-variable](http://doc.endlessparentheses.com/Fun/add-dir-local-variable) 创建局部配置



添加以下配置：

```lua
((go-mode . ((eval . (setenv "GOPACKAGESDRIVER" (concat (getenv "GOPATH") "/bin/rgopackagesdriver"))))))
```



#### 配置 rgo_lsp_server

添加以下配置：

```lua
(use-package lsp-mode
  :ensure t
  :hook (go-mode . lsp)
  :config
  (setq lsp-keymap-prefix "C-c l")  

  ;; register rgo_lsp_server
  (lsp-register-client
   (make-lsp-client
    :new-connection (lsp-stdio-connection 
                     (list (concat (getenv "GOPATH") "/bin/rgo_lsp_server")))
    :major-modes '(go-mode)
    :server-id 'rgo-lsp)))


```



#### 配置通知事件

在 Emacs 中，我们需要定义一个函数来重新启动 LSP 服务器。

```lua
(defun my-restart-go-language-server ()
  "Restart the Go language server."
  (interactive)
  (lsp-restart-workspace))

(lsp-register-custom-notification-handler
 "custom/rgo/restart_language_server"
 #'my-restart-go-language-server)

```



除此之外，还可以定义消息通知函数

```lua
(defun my-show-window-message (params)
  "Show a window message from custom notification."
  (let ((message (if (hash-table-p params)
                     (json-encode params)
                   params)))
    (message "%s" message)))  ; show in Emacs mini-buffer

(lsp-register-custom-notification-handler
 "custom/rgo/window_show"
 #'my-show-window-message)

```

#### 效果展示

![show_emacs.png](./doc/show_emacs.png)
