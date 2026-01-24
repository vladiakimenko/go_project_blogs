package service

import (
	"context"
	"errors"
	"log"

	"blog-api/internal/model"
	"blog-api/internal/repository"
	"blog-api/pkg/exception"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

func (s *CommentService) Create(
	ctx context.Context,
	userID int,
	postID int,
	req *model.CommentCreateRequest,
) (*model.Comment, *exception.ApiError) {
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		log.Printf("failed to check post existence for post_id=%d: %v", postID, err)
		return nil, exception.DatabaseError(err.Error())
	}
	if !exists {
		log.Printf("post with post_id=%d does not exist", postID)
		return nil, exception.NotFoundError("Post not found")
	}
	comment := &model.Comment{
		Content:  req.Content,
		PostID:   postID,
		AuthorID: userID,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		log.Printf("failed to create comment for post_id=%d, user_id=%d: %v", postID, userID, err)
		return nil, exception.DatabaseError(err.Error())
	}
	return comment, nil
}

func (s *CommentService) GetByID(ctx context.Context, id int) (*model.Comment, *exception.ApiError) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			log.Printf("comment with id=%d not found", id)
			return nil, exception.NotFoundError("Comment not found")
		}
		log.Printf("failed to get comment by id=%d: %v", id, err)
		return nil, exception.DatabaseError(err.Error())
	}
	return comment, nil
}

func (s *CommentService) GetByPost(
	ctx context.Context,
	postID int,
	pagination *model.PaginationParams,
) ([]*model.Comment, int, *exception.ApiError) {
	comments, err := s.commentRepo.GetByPostID(ctx, postID, *pagination.Limit, *pagination.Offset)
	if err != nil {
		log.Printf("failed to fetch comments for post_id=%d: %v", postID, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}
	total, err := s.commentRepo.GetCountByPostID(ctx, postID)
	if err != nil {
		log.Printf("failed to count comments for post_id=%d: %v", postID, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}
	return comments, total, nil
}

func (s *CommentService) Update(
	ctx context.Context,
	id int,
	userID int,
	req *model.CommentUpdateRequest,
) (*model.Comment, *exception.ApiError) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			log.Printf("comment with id=%d not found", id)
			return nil, exception.NotFoundError("Comment not found")
		}
		log.Printf("failed to get comment by id=%d: %v", id, err)
		return nil, exception.DatabaseError(err.Error())
	}
	if err := s.checkOwner(comment, userID); err != nil {
		return nil, err
	}
	comment.Content = req.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		log.Printf("failed to update comment id=%d: %v", id, err)
		return nil, exception.DatabaseError(err.Error())
	}

	return comment, nil
}

func (s *CommentService) Delete(ctx context.Context, id int, userID int) *exception.ApiError {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			log.Printf("comment with id=%d not found", id)
			return exception.NotFoundError("Comment not found")
		}
		log.Printf("failed to get comment by id=%d: %v", id, err)
		return exception.DatabaseError(err.Error())
	}
	if err := s.checkOwner(comment, userID); err != nil {
		return err
	}
	if err := s.commentRepo.Delete(ctx, id); err != nil {
		log.Printf("failed to delete comment id=%d: %v", id, err)
		return exception.DatabaseError(err.Error())
	}
	return nil
}

func (s *CommentService) GetByAuthor(
	ctx context.Context,
	authorID int,
	limit int,
	offset int,
) ([]*model.Comment, int, *exception.ApiError) {
	comments, err := s.commentRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		log.Printf("failed to fetch comments for author_id=%d: %v", authorID, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}
	total, err := s.commentRepo.GetCountByAuthorID(ctx, authorID)
	if err != nil {
		log.Printf("failed to count comments for author_id=%d: %v", authorID, err)
		return nil, 0, exception.DatabaseError(err.Error())
	}
	return comments, total, nil
}

func (s *CommentService) checkOwner(comment *model.Comment, userID int) *exception.ApiError {
	if comment.AuthorID != userID {
		log.Printf(
			"user_id=%d is not the author of comment_id=%d (author_id=%d)",
			userID,
			comment.ID,
			comment.AuthorID,
		)
		return exception.ForbiddenError("Access to the comment is forbidden")
	}
	return nil
}
