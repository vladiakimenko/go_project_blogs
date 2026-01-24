package service

import (
	"context"
	"errors"
	"log"

	"blog-api/internal/model"
	"blog-api/internal/repository"
	"blog-api/pkg/exception"
)

type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostService(
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *PostService) Create(ctx context.Context, userID int, req *model.PostCreateRequest) (*model.Post, *exception.ApiError) {
	post := &model.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		log.Printf("failed to create post for user_id=%d: %v", userID, err)
		return nil, exception.DatabaseError(err.Error())
	}
	return post, nil
}

func (s *PostService) GetByID(ctx context.Context, id int) (*model.Post, *exception.ApiError) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			log.Printf("post with id=%d not found", id)
			return nil, exception.NotFoundError("Post not found")
		}
		log.Printf("failed to get post by id=%d: %v", id, err)
		return nil, exception.DatabaseError(err.Error())
	}
	return post, nil
}

func (s *PostService) GetAll(
	ctx context.Context,
	pagination *model.PaginationParams,
) ([]*model.Post, int, *exception.ApiError) {
	posts, err := s.postRepo.GetAll(ctx, *pagination.Limit, *pagination.Offset)
	if err != nil {
		log.Printf("failed to fetch posts with limit=%d offset=%d: %v", *pagination.Limit, *pagination.Offset, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}
	total, err := s.postRepo.GetTotalCount(ctx)
	if err != nil {
		log.Printf("failed to count total posts: %v", err)
		return nil, 0, exception.DatabaseError(err.Error())
	}
	return posts, total, nil
}

func (s *PostService) Update(
	ctx context.Context,
	id int,
	userID int,
	req *model.PostUpdateRequest,
) (*model.Post, *exception.ApiError) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			log.Printf("post with id=%d not found", id)
			return nil, exception.NotFoundError("Post not found")
		}
		log.Printf("failed to fetch post id=%d: %v", id, err)
		return nil, exception.DatabaseError(err.Error())
	}
	if err := s.checkPostOwner(post, userID); err != nil {
		return nil, err
	}
	updated := false
	if req.Title != nil && *req.Title != post.Title {
		post.Title = *req.Title
		updated = true
	}
	if req.Content != nil && *req.Content != post.Content {
		post.Content = *req.Content
		updated = true
	}
	if !updated {
		return post, nil
	}
	if err := s.postRepo.Update(ctx, post); err != nil {
		log.Printf("failed to update post id=%d: %v", id, err)
		return nil, exception.DatabaseError(err.Error())
	}
	return post, nil
}

func (s *PostService) Delete(ctx context.Context, id int, userID int) *exception.ApiError {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPostNotFound) {
			log.Printf("post with id=%d not found", id)
			return exception.NotFoundError("Post not found")
		}
		log.Printf("failed to get post by id=%d: %v", id, err)
		return exception.DatabaseError(err.Error())
	}
	if err := s.checkPostOwner(post, userID); err != nil {
		return err
	}
	if err := s.postRepo.Delete(ctx, id); err != nil {
		log.Printf("failed to delete post id=%d: %v", id, err)
		return exception.DatabaseError(err.Error())
	}
	return nil
}

func (s *PostService) GetByAuthor(
	ctx context.Context,
	authorID int,
	pagination *model.PaginationParams,
) ([]*model.Post, int, *exception.ApiError) {

	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, *pagination.Limit, *pagination.Offset)
	if err != nil {
		log.Printf("failed to fetch posts for author_id=%d: %v", authorID, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}

	total, err := s.postRepo.GetCountByAuthorID(ctx, authorID)
	if err != nil {
		log.Printf("failed to count posts for author_id=%d: %v", authorID, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}

	return posts, total, nil
}

func (s *PostService) checkPostOwner(post *model.Post, userID int) *exception.ApiError {
	if post.AuthorID != userID {
		log.Printf("user_id=%d is not the author of post_id=%d", userID, post.ID)
		return exception.ForbiddenError("Access to the post is forbidden")
	}
	return nil
}
