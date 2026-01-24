package handler

import (
	"blog-api/internal/model"
	"blog-api/internal/service"
	"blog-api/pkg/exception"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// GET /api/posts?limit=10&offset=0
func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	pagination, ok := getPaginationParams(r)
	if !ok {
		exception.WriteApiError(w, exception.BadRequestError("Invalid pagination parameters"))
		return
	}

	result, total, apiErr := h.postService.GetAll(r.Context(), pagination)
	if apiErr != nil {
		exception.WriteApiError(w, apiErr)
		return
	}

	writePaginatedJSON(w, http.StatusOK, result, pagination, total)
}

// GET /api/posts/{postID}
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		exception.WriteApiError(w, exception.BadRequestError("Invalid post ID"))
		return
	}

	result, apiErr := h.postService.GetByID(r.Context(), postID)
	if apiErr != nil {
		exception.WriteApiError(w, apiErr)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// POST /api/posts
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, ok := getParsedBody[model.PostCreateRequest](r)
	if !ok {
		exception.WriteApiError(w, exception.BadRequestError("Invalid request body"))
		return
	}

	actorID, ok := getActorID(r.Context())
	if !ok {
		exception.WriteApiError(w, exception.InternalServerError("Auth misconfigured"))
		return
	}

	result, apiErr := h.postService.Create(r.Context(), actorID, body)
	if apiErr != nil {
		exception.WriteApiError(w, apiErr)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// PUT /api/posts/{postID}
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	body, ok := getParsedBody[model.PostUpdateRequest](r)
	if !ok {
		exception.WriteApiError(w, exception.BadRequestError("Invalid request body"))
		return
	}

	actorID, ok := getActorID(r.Context())
	if !ok {
		exception.WriteApiError(w, exception.InternalServerError("Auth misconfigured"))
		return
	}

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		exception.WriteApiError(w, exception.BadRequestError("Invalid post ID"))
		return
	}

	result, apiErr := h.postService.Update(r.Context(), postID, actorID, body)
	if apiErr != nil {
		exception.WriteApiError(w, apiErr)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DELETE /api/posts/{postID}
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actorID, ok := getActorID(r.Context())
	if !ok {
		exception.WriteApiError(w, exception.InternalServerError("Auth misconfigured"))
		return
	}

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		exception.WriteApiError(w, exception.BadRequestError("Invalid post ID"))
		return
	}

	apiErr := h.postService.Delete(r.Context(), postID, actorID)
	if apiErr != nil {
		exception.WriteApiError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/posts?limit=10&offset=0&author={authorID}
func (h *PostHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	pagination, ok := getPaginationParams(r)
	if !ok {
		exception.WriteApiError(w, exception.BadRequestError("Invalid pagination parameters"))
		return
	}

	authorIDStr := r.URL.Query().Get("author")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil {
		exception.WriteApiError(w, exception.BadRequestError("Invalid author ID"))
		return
	}

	result, total, apiErr := h.postService.GetByAuthor(r.Context(), authorID, pagination)
	if apiErr != nil {
		exception.WriteApiError(w, apiErr)
		return
	}

	writePaginatedJSON(w, http.StatusOK, result, pagination, total)
}
