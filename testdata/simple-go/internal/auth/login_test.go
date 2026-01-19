package auth

import "testing"

func TestLogin_Success(t *testing.T) {
    req := LoginRequest{
        Email:    "test@example.com",
        Password: "password123",
    }

    resp, err := Login(req)
    if err != nil {
        t.Fatalf("Login failed: %v", err)
    }

    if resp.User.Email != req.Email {
        t.Errorf("Email = %s; want %s", resp.User.Email, req.Email)
    }

    if resp.Token == "" {
        t.Error("Token should not be empty")
    }
}

func TestLogin_EmptyEmail(t *testing.T) {
    req := LoginRequest{
        Email:    "",
        Password: "password123",
    }

    _, err := Login(req)
    if err != ErrInvalidCredentials {
        t.Errorf("Error = %v; want ErrInvalidCredentials", err)
    }
}

func TestLogout(t *testing.T) {
    err := Logout("valid-token")
    if err != nil {
        t.Errorf("Logout failed: %v", err)
    }
}
