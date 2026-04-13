package handler

import "github.com/SawitProRecruitment/UserService/repository"

type Server struct {
	Repository repository.RepositoryInterface
	JWTSecret  string
}

type NewServerOptions struct {
	Repository repository.RepositoryInterface
	JWTSecret  string
}

func NewServer(opts NewServerOptions) *Server {
	return &Server{
		Repository: opts.Repository,
		JWTSecret:  opts.JWTSecret,
	}
}
