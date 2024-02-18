package user

type BaseResp struct {
	StatusCode    int64
	StatusMessage string
	ServiceTime   int64
}

type User struct {
	UserId   int64
	Username string
	Password string
}

type CreateUserRequest struct {
	Username string
	Password string
}

type CreateUserResponse struct {
	BaseResp *BaseResp
}

type MGetUserRequest struct {
	UserIds []int64
}

type MGetUserResponse struct {
	Users    []*User
	BaseResp *BaseResp
}
