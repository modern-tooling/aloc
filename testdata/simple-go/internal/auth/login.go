package auth

import (
    "errors"
    "time"
)

type User struct {
    ID        string
    Email     string
    CreatedAt time.Time
}

type LoginRequest struct {
    Email    string
    Password string
}

type LoginResponse struct {
    User  *User
    Token string
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func Login(req LoginRequest) (*LoginResponse, error) {
    if req.Email == "" || req.Password == "" {
        return nil, ErrInvalidCredentials
    }

    user := &User{
        ID:        "user-123",
        Email:     req.Email,
        CreatedAt: time.Now(),
    }

    return &LoginResponse{
        User:  user,
        Token: "token-abc-123",
    }, nil
}

func Logout(token string) error {
    if token == "" {
        return errors.New("invalid token")
    }
    return nil
}
