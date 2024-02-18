package main

import (
	"github.com/cloudwego-contrib/rgo/examples/server/rgo"

	// 注意的点：需要匿名引一下
	// todo 强制写到 main 里
	_ "github.com/cloudwego-contrib/rgo/examples/server/service"
)

func main() {
	println("init logic")
	rgo.Run()
	println("end logic")
}
