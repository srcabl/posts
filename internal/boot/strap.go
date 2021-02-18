package boot

import (
	"github.com/pkg/errors"
	"github.com/srcabl/posts/internal/server"
	"github.com/srcabl/posts/internal/service"
	pb "github.com/srcabl/protos/posts"
	"github.com/srcabl/services/pkg/config"
	"github.com/srcabl/services/pkg/db/mysql"
	"google.golang.org/grpc"
)

// Strap initializes the user service
type Strap struct {
	Config     *config.Service
	Middleware grpc.ServerOption
	Service    pb.PostsServiceServer
	Server     server.GRPC

	onconnect  map[string](func() (func() error, error))
	onshutdown map[string](func() error)
}

// New news up boot and all application services
func New(cfg *config.Service) (*Strap, error) {
	db, err := mysql.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed new db client")
	}

	middleware := grpc.EmptyServerOption{}

	srvc, err := service.New(db)
	if err != nil {
		return nil, err
	}

	srv, err := server.New(cfg, middleware, srvc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to new server")
	}

	return &Strap{
		Config:     cfg,
		Middleware: middleware,
		Service:    srvc,
		Server:     srv,

		onconnect: map[string](func() (func() error, error)){
			"database connection": db.Connect,
			"service run":         srv.Run,
		},
		onshutdown: map[string](func() error){},
	}, nil
}

// Connect connects all application services
func (s *Strap) Connect() error {
	for name, connect := range s.onconnect {
		os, err := connect()
		if err != nil {
			return errors.Wrapf(err, "%s failed", name)
		}
		s.onshutdown[name] = os
	}
	return nil
}

// Shutdown shuts down all application srvices
func (s *Strap) Shutdown() []error {
	var errs []error
	for name, shutdown := range s.onshutdown {
		err := shutdown()
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "%s failed", name))
		}
	}
	return errs
}
