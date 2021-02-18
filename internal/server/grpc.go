package server

import (
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
	pb "github.com/srcabl/protos/posts"
	"github.com/srcabl/services/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPC defines the actions of a grpc server
type GRPC interface {
	Run() (func() error, error)
}

// GRPCServer is the sources grpc server
type GRPCServer struct {
	address string
	port    int
	server  *grpc.Server
}

// New news up a users grpc server
func New(config *config.Service, middleware grpc.ServerOption, service pb.PostsServiceServer) (GRPC, error) {
	server := grpc.NewServer(middleware)
	pb.RegisterPostsServiceServer(server, service)
	reflection.Register(server)
	return &GRPCServer{
		server:  server,
		address: config.Server.Address,
		port:    config.Server.Port,
	}, nil
}

// Run starts the grpc server
func (s *GRPCServer) Run() (func() error, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.address, s.port))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to begin listening on port %d", s.port)
	}
	log.Printf("Serving Posts on post: %d\n", s.port)
	if err := s.server.Serve(lis); err != nil {
		return nil, errors.Wrapf(err, "failed to serve on port %d", s.port)
	}
	return lis.Close, nil
}
