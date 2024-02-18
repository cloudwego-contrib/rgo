# RGO
RGO 目前处于 MVP 阶段

# 运行步骤

## 安装 RGO

```shell
go install github.com/cloudwego-contrib/rgo@latest
```

## 修改 MySQL 地址

目前支持通过环境变量设置 MySQL 地址，使用内置 docker 可忽略此步骤。

```shell
export RGO_MYSQL_DSN="gorm:gorm@tcp(localhost:3306)/gorm?charset=utf8&parseTime=True&loc=Local"
```

## 启动 MySQL docker 服务

```shell
docker-compose up
```

## 修改 MySQL 地址

目前支持通过环境变量设置 MySQL 地址，使用上述 docker 命令可忽略此步骤。

```shell
export RGO_MYSQL_DSN="gorm:gorm@tcp(localhost:3306)/gorm?charset=utf8&parseTime=True&loc=Local"
```
## 准备公共结构体

- 目前准备了一份远程测试结构体 `github.com/chaoranz758/rgo_struct` 。
- 目前仅支持远程结构体，本地结构体尚未支持。

## 启动 Server Example

```shell
cd examples/server
go run -toolexec rgo main.go
```

效果:
```shell
init logic
2024/02/19 03:36:17.364462 server.go:83: [Info] KITEX: server listen at addr=127.0.0.1:8888
create service success, username: xiaoming, password: 123456
```

## 启动 Client Example

```shell
cd examples/client
go run -toolexec rgo main.go
```

效果:
```shell
create service resp: &{1000 success 1708285124}
```
