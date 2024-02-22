package main

import (
	"github.com/cloudwego-contrib/rgo/cmd/rgo"

	"github.com/cloudwego-contrib/rgo/examples/server/service"
)

func main() {
	println("init logic")
	rgo.Run("user_service", service.CreateUser, service.MGetUser)
	println("end logic")
}
