package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/srcabl/services/pkg/db/mysql"
	"github.com/srcabl/services/pkg/proto"
)

// DataRepositoryGetter defines the beahvior of a data repo getter
type DataRepositoryGetter interface {
	GetPost(context.Context, string) (*DBPost, error)
	GetLinkByUUID(context.Context, string) (*DBLink, error)
	GetLinkByURL(context.Context, string) (*DBLink, error)
	GetUsersPosts(context.Context, string, *proto.PaginationToken) ([]*DBPost, []*DBLink, error)
}

// DataRepositoryCreator defines the beahvior of a data repo creator
type DataRepositoryCreator interface {
	CreateLink(context.Context, *DBLink) error
	CreatePost(context.Context, *DBPost) error
}

// DataRepository defines the behavior of a data repo
type DataRepository interface {
	DataRepositoryGetter
	DataRepositoryCreator
}

type dataRepository struct {
	db *mysql.Client
}

//NewDataRepository news up a data repository
func NewDataRepository(db *mysql.Client) (DataRepository, error) {
	return &dataRepository{
		db: db,
	}, nil
}

const getUserPostQuery = `
SELECT
	p.uuid,
	p.user_uuid,
	p.link_uuid,
	p.comment,
	p.created_by_uuid,
	p.created_at,
	p.updated_by_uuid,
	p.updated_at,
FROM
	posts p
WHERE
	p.uuid=?
`

// GetPost gets a post by uuid
func (dr *dataRepository) GetPost(ctx context.Context, uuid string) (*DBPost, error) {
	post := DBPost{}
	scanErr := dr.db.DB.QueryRowContext(ctx, getUserPostQuery, uuid).Scan(
		&post.UUID,
		&post.UserUUID,
		&post.LinkUUID,
		&post.Comment,
		&post.CreatedByUUID,
		&post.CreatedAt,
		&post.UpdatedByUUID,
		&post.UpdatedAt,
	)
	if scanErr != nil {
		return nil, errors.Wrapf(scanErr, "failed to scan a rom of posts for user %s", uuid)
	}
	return &post, nil
}

// GetLinkByUUID gets a link by the uuid
func (dr *dataRepository) GetLinkByUUID(ctx context.Context, uuid string) (*DBLink, error) {
	whereStatement := "l.uuid=?"
	link, err := dr.getLinkByParam(ctx, whereStatement, uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get link with uuid %s", uuid)
	}
	return link, nil
}

// GetLinkByURL gets a link by the url
func (dr *dataRepository) GetLinkByURL(ctx context.Context, url string) (*DBLink, error) {
	whereStatement := "l.url=?"
	link, err := dr.getLinkByParam(ctx, whereStatement, url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get link with url %s", url)
	}
	return link, nil
}

const getLinkQuery = `
SELECT
	l.uuid
	l.url,
	l.created_by_uuid,
	l.created_at,
	l.updated_by_uuid,
	l.updated_at,
	GROUP_CONCAT(lsh.source_uuid)
FROM
	links l
INNER JOIN
	link_source_heads lsh
ON
	l.uuid=lsh.link_uuid
WHERE
`

func (dr *dataRepository) getLinkByParam(ctx context.Context, whereStatement string, param string) (*DBLink, error) {
	link := DBLink{}
	var aggSources string
	query := fmt.Sprintf("%s %s", getLinkQuery, whereStatement)
	scanErr := dr.db.DB.QueryRowContext(ctx, query, param).Scan(
		&link.UUID,
		&link.URL,
		&link.CreatedByUUID,
		&link.CreatedAt,
		&link.UpdatedByUUID,
		&link.UpdatedAt,
		&aggSources,
	)
	if aggSources != "" {
		link.SourceHeadUUIDs = strings.Split(aggSources, ",")
	}
	if scanErr != nil {
		return nil, errors.Wrapf(scanErr, "failed to scan a rom of link for param %s", param)
	}
	return &link, nil
}

const getUsersPostsQuery = `
SELECT
	p.uuid,
	p.user_uuid,
	p.link_uuid,
	p.comment,
	p.created_by_uuid,
	p.created_at,
	p.updated_by_uuid,
	p.updated_at,
	l.uuid,
	l.url,
	l.created_by_uuid,
	l.created_at,
	l.updated_by_uuid,
	l.updated_at,
	(SELECT GROUP_CONCAT(lsh.source_uuid) FROM link_source_heads lsh WHERE lsh.link_uuid=l.uuid) AS link_sources
FROM
	posts p
INNER JOIN
	links l
ON
	p.link_uuid=l.uuid
WHERE
	p.user_uuid=?	
`

// GetUsersPosts gets the posts from a user from the database
func (dr *dataRepository) GetUsersPosts(ctx context.Context, userUUID string, token *proto.PaginationToken) ([]*DBPost, []*DBLink, error) {
	query := token.ApplyToQuery(getUsersPostsQuery, "p.created_at")
	fmt.Printf("query: %s\n", query)
	rows, err := dr.db.DB.QueryContext(ctx, query, userUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to query posts for user %s", userUUID)
	}
	var posts []*DBPost
	var links []*DBLink
	for rows.Next() {
		post := DBPost{}
		link := DBLink{}
		var aggSources string
		scanErr := rows.Scan(
			&post.UUID,
			&post.UserUUID,
			&post.LinkUUID,
			&post.Comment,
			&post.CreatedByUUID,
			&post.CreatedAt,
			&post.UpdatedByUUID,
			&post.UpdatedAt,
			&link.UUID,
			&link.URL,
			&link.CreatedByUUID,
			&link.CreatedAt,
			&link.UpdatedByUUID,
			&link.UpdatedAt,
			&aggSources,
		)
		if scanErr != nil {
			return nil, nil, errors.Wrapf(scanErr, "failed to scan a rom of posts for user %s", userUUID)
		}
		posts = append(posts, &post)
		if aggSources != "" {
			link.SourceHeadUUIDs = strings.Split(aggSources, ",")
		}
		links = append(links, &link)
	}
	return posts, links, nil
}

const createPostStatement = `
INSERT INTO
	posts (
		uuid,
		user_uuid,
		link_uuid,
		title,
		comment,
		created_by_uuid,
		created_at,
		updated_by_uuid,
		updated_at
	)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?)
`

// CreatePost adds a post in the database
func (dr *dataRepository) CreatePost(ctx context.Context, post *DBPost) error {
	tx, err := dr.db.DB.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("Failed to begin tx: %+v\n", err)
		return errors.Wrapf(err, "failed to begin transaction")
	}
	stm, err := tx.PrepareContext(ctx, createPostStatement)
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			fmt.Printf("Failed to rollback after prepare: %+v\n", rollErr)
			return errors.Wrapf(rollErr, "failed to rollback after failing to create post %+v", post)
		}
		fmt.Printf("Failed to prepare statement: %+v\n", err)
		return errors.Wrapf(err, "failed to prepare statement to create post %+v", post)
	}
	_, err = stm.ExecContext(ctx,
		post.UUID,
		post.UserUUID,
		post.LinkUUID,
		post.Title,
		post.Comment,
		post.CreatedByUUID,
		post.CreatedAt,
		post.UpdatedByUUID.String,
		post.UpdatedAt.Int64,
	)
	if err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			fmt.Printf("Failed to rollback after execute statement: %+v\n", rollErr)
			return errors.Wrapf(rollErr, "failed to rollback after failing to create post %+v", post)
		}
		fmt.Printf("Failed to execute statement: %+v\n", err)
		return errors.Wrapf(err, "failed to execute statment to create post %+v", post)
	}
	if err := tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			fmt.Printf("Failed to rollback after commit: %+v\n", rollErr)
			return errors.Wrapf(rollErr, "failed to rollback after failing to create post %+v", post)
		}
		fmt.Printf("Failed to commit: %+v\n", err)
		return errors.Wrapf(err, "failed to create post %+v", post)
	}
	return nil
}

// CreateLink adds a link in the database
func (dr *dataRepository) CreateLink(ctx context.Context, link *DBLink) error {
	tx, err := dr.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to begin transaction")
	}
	if err := dr.createLink(ctx, tx, link); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to create link %+v", link)
		}
		return errors.Wrap(err, "failed to create in the link table")
	}
	if err := dr.createLinkSourceHeads(ctx, tx, link); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to create link %+v", link)
		}
		return errors.Wrap(err, "failed to create in the link source head table")
	}
	if err := tx.Commit(); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Wrapf(rollErr, "failed to rollback after failing to create link %+v", link)
		}
		return errors.Wrapf(err, "failed to create link %+v", link)
	}
	return nil
}

const createLinkStatement = `
INSERT INTO
	links (
		uuid,
		url,
		created_by_uuid,
		created_at,
		updated_by_uuid,
		updated_at
	)
VALUES
	(?, ?, ?, ?, ?, ?)
`

func (dr *dataRepository) createLink(ctx context.Context, tx *sql.Tx, link *DBLink) error {
	stm, err := tx.PrepareContext(ctx, createLinkStatement)
	if err != nil {
		return errors.Wrapf(err, "failed to prepare statement to create link %+v", link)
	}
	_, err = stm.ExecContext(ctx,
		link.UUID,
		link.URL,
		link.CreatedByUUID,
		link.CreatedAt,
		link.UpdatedByUUID.String,
		link.UpdatedAt.Int64,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to execute statment to create link %+v", link)
	}
	return nil
}

const createLinkSourceHeadStatement = `
INSERT INTO
	link_source_heads(
		link_uuid,
		source_uuid
	)
VALUES
	(?, ?)
`

func (dr *dataRepository) createLinkSourceHeads(ctx context.Context, tx *sql.Tx, link *DBLink) error {
	stm, err := tx.PrepareContext(ctx, createLinkSourceHeadStatement)
	if err != nil {
		return errors.Wrapf(err, "failed to prepare statement to create link source heads %+v", link)
	}
	for _, s := range link.SourceHeadUUIDs {
		_, err = stm.ExecContext(ctx,
			link.UUID,
			s,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to execute statment to create link source head %+v", link)
		}
	}
	return nil
}
