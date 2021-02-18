package service

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	postspb "github.com/srcabl/protos/posts"
	pb "github.com/srcabl/protos/shared"
	"github.com/srcabl/services/pkg/proto"
)

// DBLink is the database link model
type DBLink struct {
	UUID            string
	URL             string
	SourceHeadUUIDs []string
	CreatedByUUID   string
	CreatedAt       int64
	UpdatedByUUID   sql.NullString
	UpdatedAt       sql.NullInt64
}

// CreatedByUUIDString satisfies the services helper to transform db auditfields to grpc auditfields
func (l *DBLink) CreatedByUUIDString() string {
	return l.CreatedByUUID
}

// CreatedAtUnixInt satisfies the services helper to transform db auditfields to grpc auditfields
func (l *DBLink) CreatedAtUnixInt() int64 {
	return l.CreatedAt
}

// UpdatedByUUIDNullString satisfies the services helper to transform db auditfields to grpc auditfields
func (l *DBLink) UpdatedByUUIDNullString() sql.NullString {
	return l.UpdatedByUUID
}

// UpdatedAtUnixNullInt satisfies the services helper to transform db auditfields to grpc auditfields
func (l *DBLink) UpdatedAtUnixNullInt() sql.NullInt64 {
	return l.UpdatedAt
}

// ToGRPC transforms the dbuser to proto link
func (l *DBLink) ToGRPC() (*pb.Link, error) {
	id, err := uuid.FromString(l.UUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to transform uuid: %s", l.UUID)
	}
	var srcUUIDs [][]byte
	for _, s := range l.SourceHeadUUIDs {
		sUUID, err := uuid.FromString(s)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to transform source uuid : %s", s)
		}
		srcUUIDs = append(srcUUIDs, sUUID.Bytes())
	}
	auditFields, err := proto.DBAuditFieldsToGRPC(l)
	if err != nil {
		return nil, errors.Wrap(err, "failed to transform auditfields")
	}
	return &pb.Link{
		Uuid:        id.Bytes(),
		SourceHeads: srcUUIDs,
		Url:         l.URL,
		AuditFields: auditFields,
	}, nil
}

// HydrateLinkModelForCreate creates a db post from a proto post and fills in any missing data
func HydrateLinkModelForCreate(req *postspb.CreateLinkRequest) (*DBLink, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate uuid for link")
	}
	var sourceHeadUUIDs []string
	for _, s := range req.SourceHeadUuids {
		sUUID, err := uuid.FromBytes(s)
		if err != nil {
			return nil, errors.Wrap(err, "failed to transform uuid for source head")
		}
		sourceHeadUUIDs = append(sourceHeadUUIDs, sUUID.String())
	}
	now := time.Now().Unix()
	return &DBLink{
		UUID:            newUUID.String(),
		URL:             req.Url,
		SourceHeadUUIDs: sourceHeadUUIDs,
		CreatedByUUID:   newUUID.String(),
		CreatedAt:       now,
		UpdatedByUUID:   sql.NullString{Valid: true, String: newUUID.String()},
		UpdatedAt:       sql.NullInt64{Valid: true, Int64: now},
	}, nil
}
