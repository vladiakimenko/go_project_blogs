package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"blog-api/internal/model"
	"blog-api/pkg/database"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

type PostRepo struct {
	db *database.DatabaseManager
}

func NewPostRepo(db *database.DatabaseManager) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	result := r.db.ORM.WithContext(ctx).Create(post)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *PostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {
	var post model.Post
	err := r.db.ORM.WithContext(ctx).First(&post, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}
	return &post, nil
}

func (r *PostRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	var posts []*model.Post
	err := r.db.ORM.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	return posts, nil
}

func (r *PostRepo) GetTotalCount(ctx context.Context) (int, error) {
	var count int64
	err := r.db.ORM.WithContext(ctx).Model(&model.Post{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}
	return int(count), nil
}

func (r *PostRepo) Update(ctx context.Context, post *model.Post) error {
	post.UpdatedAt = time.Now()

	result := r.db.ORM.WithContext(ctx).Model(&model.Post{}).Where("id = ?", post.ID).Updates(
		map[string]any{
			"title":      post.Title,
			"content":    post.Content,
			"updated_at": post.UpdatedAt,
		},
	)
	if result.Error != nil {
		return fmt.Errorf("failed to update post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *PostRepo) Delete(ctx context.Context, id int) error {
	result := r.db.ORM.WithContext(ctx).Delete(&model.Post{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete post: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *PostRepo) Exists(ctx context.Context, id int) (bool, error) {
	var count int64
	err := r.db.ORM.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}
	return count > 0, nil
}

func (r *PostRepo) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	var posts []*model.Post
	err := r.db.ORM.WithContext(ctx).Where("author_id = ?", authorID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by author ID %d: %w", authorID, err)
	}
	return posts, nil
}

func (r *PostRepo) GetCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	var count int64
	err := r.db.ORM.WithContext(ctx).Model(&model.Post{}).Where("author_id = ?", authorID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count posts by author ID %d: %w", authorID, err)
	}
	return int(count), nil
}
