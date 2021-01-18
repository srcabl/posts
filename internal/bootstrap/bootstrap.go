package bootstrap

import (
	"github.com/srcabl/posts/internal/config"
	"github.com/srcabl/posts/internal/service"
	pb "github.com/srcabl/protos/posts"
)

type Bootstrap struct {
	config  *config.Environment
	service pb.PostsServiceServer
}

func New(cfg *config.Environment) (*Bootstrap, error) {

	srvc, err := service.New()
	if err != nil {
		return nil, err
	}

	return &Bootstrap{
		config:  cfg,
		service: srvc,
	}, nil
}
