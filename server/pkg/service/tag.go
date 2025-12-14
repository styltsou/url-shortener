package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/styltsou/url-shortener/server/pkg/db"
	apperrors "github.com/styltsou/url-shortener/server/pkg/errors"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

type TagQueries interface {
	ListUserTags(ctx context.Context, userID string) ([]db.ListUserTagsRow, error)
	CreateTag(ctx context.Context, arg db.CreateTagParams) (db.CreateTagRow, error)
	UpdateTag(ctx context.Context, arg db.UpdateTagParams) (db.UpdateTagRow, error)
	DeleteTag(ctx context.Context, arg db.DeleteTagParams) (db.DeleteTagRow, error)
	DeleteTags(ctx context.Context, arg db.DeleteTagsParams) ([]db.DeleteTagsRow, error)
}

type TagService struct {
	queries TagQueries
	logger  logger.Logger
}

func NewTagService(queries TagQueries, logger logger.Logger) *TagService {
	return &TagService{
		queries: queries,
		logger:  logger,
	}
}

func (s *TagService) ListAllTags(ctx context.Context, userID string) ([]db.ListUserTagsRow, error) {
	s.logger.Debug("Querying database for user tags",
		zap.String("user_id", userID),
	)

	tags, err := s.queries.ListUserTags(ctx, userID)
	if err != nil {
		s.logger.Error("Database query failed for ListUserTags",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	s.logger.Debug("Database query completed for ListUserTags",
		zap.String("user_id", userID),
		zap.Int("tags_found", len(tags)),
	)

	return tags, nil
}

func (s *TagService) CreateTag(ctx context.Context, userID string, name string) (db.CreateTagRow, error) {
	createdTag, err := s.queries.CreateTag(ctx, db.CreateTagParams{
		Name:   name,
		UserID: userID,
	})

	if err != nil {
		// Check for unique constraint violation (tag name already exists for this user)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.CreateTagRow{},
				fmt.Errorf("%w: tag name '%s' already exists", apperrors.TagNameTaken, name)
		}

		return db.CreateTagRow{}, fmt.Errorf("failed to create tag: %w", err)
	}

	return createdTag, nil
}

func (s *TagService) UpdateTag(ctx context.Context, userID string, tagID uuid.UUID, name string) (db.UpdateTagRow, error) {
	updatedTag, err := s.queries.UpdateTag(ctx, db.UpdateTagParams{
		Name:   name,
		ID:     tagID,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateTagRow{},
				fmt.Errorf("%w: %v", apperrors.TagNotFound, err)
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.UpdateTagRow{},
				fmt.Errorf("%w: tag name '%s' already exists", apperrors.TagNameTaken, name)
		}

		return db.UpdateTagRow{}, fmt.Errorf("failed to update tag: %w", err)
	}

	return updatedTag, nil
}

func (s *TagService) DeleteTag(ctx context.Context, userID string, tagID uuid.UUID) (db.DeleteTagRow, error) {
	deletedTag, err := s.queries.DeleteTag(ctx, db.DeleteTagParams{
		ID:     tagID,
		UserID: userID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.DeleteTagRow{}, fmt.Errorf("%w: %v", apperrors.TagNotFound, err)
		}

		return db.DeleteTagRow{}, fmt.Errorf("failed to delete tag: %w", err)
	}

	return deletedTag, nil
}

func (s *TagService) DeleteTags(ctx context.Context, userID string, tagIDs []uuid.UUID) ([]db.DeleteTagsRow, error) {
	if len(tagIDs) == 0 {
		return []db.DeleteTagsRow{}, nil
	}

	deletedTags, err := s.queries.DeleteTags(ctx, db.DeleteTagsParams{
		TagIDs: tagIDs,
		UserID: userID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to delete tags: %w", err)
	}

	return deletedTags, nil
}
