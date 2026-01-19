package api

import (
    "encoding/json"
    "net/http"
)

type Server struct {
    addr string
}

func NewServer(addr string) *Server {
    return &Server{addr: addr}
}

func (s *Server) Start() error {
    http.HandleFunc("/health", s.handleHealth)
    http.HandleFunc("/api/v1/users", s.handleUsers)
    return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
    users := []map[string]string{
        {"id": "1", "name": "Person A"},
        {"id": "2", "name": "Person B"},
    }
    json.NewEncoder(w).Encode(users)
}
