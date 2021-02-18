package service

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	pb "github.com/srcabl/protos/posts"
	"github.com/srcabl/protos/shared"
	"github.com/srcabl/services/pkg/db/mysql"
	"github.com/srcabl/services/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements the posts service
type Handler struct {
	pb.UnimplementedPostsServiceServer
	datarepo DataRepository
}

// New creates the service handler
func New(db *mysql.Client) (*Handler, error) {
	dataRepo, err := NewDataRepository(db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create data repo")
	}
	return &Handler{
		datarepo: dataRepo,
	}, nil
}

// HealthCheck is the base healthcheck for the service
func (h *Handler) HealthCheck(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {

	return nil, nil
}

// GetPost gets a post
func (h *Handler) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	postID, err := uuid.FromBytes(req.PostUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrapf(err, "failed to convert uuid").Error())
	}
	dbPost, err := h.datarepo.GetPost(ctx, postID.String())
	if err != nil {
		return nil, status.Error(codes.Internal, "something happened")
	}
	pbPost, err := dbPost.ToGRPC()
	if err != nil {
		return nil, status.Error(codes.Internal, "something happened")
	}
	return &pb.GetPostResponse{Post: pbPost}, nil
}

// GetLink gets a link
func (h *Handler) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	var dbLink *DBLink
	if req.GetBy == pb.GetLinkRequest_URL {
		l, err := h.datarepo.GetLinkByURL(ctx, req.Url)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "Something happened").Error())
		}
		dbLink = l
	}
	if req.GetBy == pb.GetLinkRequest_UUID {
		l, err := h.datarepo.GetLinkByUUID(ctx, string(req.LinkUuid))
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "Something happened").Error())
		}
		dbLink = l
	}
	if dbLink == nil {
		return nil, status.Error(codes.InvalidArgument, "UNKNOWN get by value is not supported")
	}
	pbLink, err := dbLink.ToGRPC()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "something happened").Error())
	}
	return &pb.GetLinkResponse{Link: pbLink}, nil
}

// ListUsersPosts gets list of posts
func (h *Handler) ListUsersPosts(ctx context.Context, req *pb.ListUsersPostsRequest) (*pb.ListUsersPostsResponse, error) {
	userID, err := uuid.FromBytes(req.UserUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrapf(err, "failed to convert uuid").Error())
	}
	pagToken := proto.NewTokenFromRequest(req)
	fmt.Printf("Pag Token: %+v\n", pagToken)
	dbPosts, dbLinks, err := h.datarepo.GetUsersPosts(ctx, userID.String(), pagToken)
	var posts []*shared.Post
	var links []*shared.Link
	for i, dbp := range dbPosts {
		p, err := dbp.ToGRPC()
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to transform dbpost").Error())
		}
		l, err := dbLinks[i].ToGRPC()
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to transform dblink").Error())
		}
		posts = append(posts, p)
		links = append(links, l)
	}
	nextToken, err := pagToken.EncodeNextToken()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get next token").Error())
	}
	return &pb.ListUsersPostsResponse{
		Posts:         posts,
		Links:         links,
		NextPageToken: nextToken,
	}, nil
}

// CreateLink is the handler for creating posts
func (h *Handler) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.CreateLinkResponse, error) {
	dbLink, err := HydrateLinkModelForCreate(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to hydrate link for create").Error())
	}
	if err := h.datarepo.CreateLink(ctx, dbLink); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to create link").Error())
	}
	hydratedPBLink, err := dbLink.ToGRPC()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "something ridiculos").Error())
	}
	return &pb.CreateLinkResponse{
		Link: hydratedPBLink,
	}, nil
}

// CreatePost is the handler for creating posts
func (h *Handler) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	fmt.Print("Creating Post: %+v\n")
	dbPost, err := HydratePostModelForCreate(req)
	if err != nil {
		fmt.Printf("Failed to hydrate: %+v\n", err)
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to hydrate post for create").Error())
	}
	if err := h.datarepo.CreatePost(ctx, dbPost); err != nil {
		fmt.Printf("Failed to create post: %+v\n", err)
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "failed to create link").Error())
	}
	hydratedPBPost, err := dbPost.ToGRPC()
	if err != nil {
		fmt.Printf("Failed to grpc: %+v\n", err)
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "something ridiculos").Error())
	}
	fmt.Printf("Returning from Create Post\n")
	return &pb.CreatePostResponse{
		Post: hydratedPBPost,
	}, nil
}

// DeletePost deletes a post
func (h *Handler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {

	return nil, errors.New("not implemented")
}
