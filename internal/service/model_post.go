package service

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	postspb "github.com/srcabl/protos/posts"
	sharedpb "github.com/srcabl/protos/shared"
	"github.com/srcabl/services/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DBPost is the database model of a post
type DBPost struct {
	UUID          string
	UserUUID      string
	LinkUUID      string
	Title         string
	Comment       string
	CreatedByUUID string
	CreatedAt     int64
	UpdatedByUUID sql.NullString
	UpdatedAt     sql.NullInt64
}

// CreatedByUUIDString satisfies the services helper to transform db auditfields to grpc auditfields
func (p *DBPost) CreatedByUUIDString() string {
	return p.CreatedByUUID
}

// CreatedAtUnixInt satisfies the services helper to transform db auditfields to grpc auditfields
func (p *DBPost) CreatedAtUnixInt() int64 {
	return p.CreatedAt
}

// UpdatedByUUIDNullString satisfies the services helper to transform db auditfields to grpc auditfields
func (p *DBPost) UpdatedByUUIDNullString() sql.NullString {
	return p.UpdatedByUUID
}

// UpdatedAtUnixNullInt satisfies the services helper to transform db auditfields to grpc auditfields
func (p *DBPost) UpdatedAtUnixNullInt() sql.NullInt64 {
	return p.UpdatedAt
}

// ToGRPC transforms the dbuser to proto user
func (p *DBPost) ToGRPC() (*sharedpb.Post, error) {
	id, err := uuid.FromString(p.UUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform uuid: %s", p.UUID)
	}
	userid, err := uuid.FromString(p.UserUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform user uuid: %s", p.UUID)
	}
	linkid, err := uuid.FromString(p.LinkUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform link uuid: %s", p.UUID)
	}
	auditFields, err := proto.DBAuditFieldsToGRPC(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to transform auditfields")
	}
	return &sharedpb.Post{
		Uuid:     id.Bytes(),
		UserUuid: userid.Bytes(),
		LinkUuid: linkid.Bytes(),
		Comment: &sharedpb.Post_PostComment{
			Title:          p.Title,
			PrimaryContent: p.Comment,
		},
		AuditFields: auditFields,
	}, nil
}

// HydratePostModelForCreate creates a db post from a proto post and fills in any missing data
func HydratePostModelForCreate(req *postspb.CreatePostRequest) (*DBPost, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to generate uuid for post").Error())
	}
	userid, err := uuid.FromBytes(req.UserUuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform user uuid: %s", req.UserUuid)
	}
	linkid, err := uuid.FromBytes(req.LinkUuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform link uuid: %s", req.LinkUuid)
	}
	now := time.Now().Unix()
	return &DBPost{
		UUID:          newUUID.String(),
		UserUUID:      userid.String(),
		LinkUUID:      linkid.String(),
		Title:         req.Title,
		Comment:       req.Comment,
		CreatedByUUID: newUUID.String(),
		CreatedAt:     now,
		UpdatedByUUID: sql.NullString{Valid: true, String: newUUID.String()},
		UpdatedAt:     sql.NullInt64{Valid: true, Int64: now},
	}, nil
}
