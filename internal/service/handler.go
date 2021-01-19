package service

import (
	"context"
	"errors"

	pb "github.com/srcabl/protos/posts"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	pb.UnimplementedPostsServiceServer
}

// New creates the service handler
func New() (*Handler, error) {
	return &Handler{}, nil
}

// HealthCheck is the base healthcheck for the service
func (h *Handler) HealthCheck(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {

	return nil, nil
}

// CreatePost is the handler for creating posts
func (h *Handler) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {

	return nil, errors.New("not implemented")
}

// StartPostsList starts the listing of posts
func (h *Handler) StartPostsList(ctx context.Context, req *pb.StartPostsListRequest) (*pb.StartPostsListResponse, error) {

	return nil, errors.New("not implemented")
}

// GetPostsList gets list of posts
func (h *Handler) GetPostsList(ctx context.Context, req *pb.GetPostsListRequest) (*pb.GetPostsListResponse, error) {

	return nil, errors.New("not implemented")
}

// DeletePost deletes a post
func (h *Handler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {

	return nil, errors.New("not implemented")
}
