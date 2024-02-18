package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego-contrib/rgo/examples/remote_struct/user"
)

// rgo:server:codehub:dns:user_service
func CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	fmt.Printf("create service success, username: %v, password: %v", req.Username, req.Password)
	return &user.CreateUserResponse{BaseResp: &user.BaseResp{
		StatusCode:    1000,
		StatusMessage: "success",
		ServiceTime:   time.Now().Unix(),
	}}, nil
}

// rgo:server:codehub:dns:user_service
func MGetUser(ctx context.Context, req *user.MGetUserRequest) (*user.MGetUserResponse, error) {
	fmt.Printf("mget service process...")
	fmt.Printf("mget service success")
	return &user.MGetUserResponse{
		Users: []*user.User{
			{
				Username: "xiaoming",
				Password: "123456",
			},
		},
		BaseResp: &user.BaseResp{
			StatusCode:    1000,
			StatusMessage: "success",
			ServiceTime:   time.Now().Unix(),
		},
	}, nil
}
