package api

import "context"

type UserService struct {
    // service implementation
}

func NewUserService() *UserService {
    return &UserService{}
}

func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
    user := &User{
        Id:    req.Id,
        Name:  "Person A",
        Email: "person@example.com",
    }
    return &GetUserResponse{User: user}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
    users := []*User{
        {Id: "1", Name: "Person A", Email: "a@example.com"},
        {Id: "2", Name: "Person B", Email: "b@example.com"},
    }
    return &ListUsersResponse{Users: users}, nil
}
