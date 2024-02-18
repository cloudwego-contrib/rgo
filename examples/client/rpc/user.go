package rpc

import (
	"context"

	"github.com/cloudwego-contrib/rgo/examples/remote_struct/user"
)

// rgo:client:codehub:dns:user_service@v1
func CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	return nil, nil
}
