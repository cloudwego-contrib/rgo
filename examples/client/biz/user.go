package biz

import (
	"context"
	"fmt"

	"github.com/chaoranz758/rgo_struct/user"
	"github.com/cloudwego-contrib/rgo/examples/client/rpc"
)

func Biz() {
	resp, err := rpc.CreateUser(context.Background(), &user.CreateUserRequest{
		Username: "xiaoming",
		Password: "123456",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("create service resp: %v", resp.BaseResp)
}
