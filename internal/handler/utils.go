package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"blog-api/internal/middleware"
	"blog-api/internal/model"
	"blog-api/pkg/exception"
	"blog-api/pkg/validator"
)

func writeJSON(
	w http.ResponseWriter,
	status int,
	v any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		exception.WriteApiError(
			w,
			exception.InternalServerError(err.Error()),
		)
	}
}

func getParsedBody[T any](
	r *http.Request,
) (*T, bool) {
	data := r.Context().Value(middleware.ParsedBodyKey).(*T)
	if err := validator.ModelValidate(data); err != nil {
		log.Println(err.Error())
		return nil, false
	}
	return data, true
}

func getActorID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		log.Println(fmt.Sprintf("%s missing in context", middleware.UserIDKey))
		return 0, false
	}
	return userID, true
}

func getPaginationParams(r *http.Request) (*model.PaginationParams, bool) {
	pagination := &model.PaginationParams{}
	query := r.URL.Query()

	if l := query.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			pagination.Limit = &v
		} else {
			log.Println("Invalid limit value:", err)
			return nil, false
		}
	}

	if o := query.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			pagination.Offset = &v
		} else {
			log.Println("Invalid offset value:", err)
			return nil, false
		}
	}

	if err := validator.ModelValidate(pagination); err != nil {
		log.Println(err.Error())
		return nil, false
	}

	return pagination, true
}

func writePaginatedJSON[T any](
	w http.ResponseWriter,
	status int,
	data T,
	pagination *model.PaginationParams,
	total int,
) {
	resp := model.PaginatedResponse[T]{
		Data:   data,
		Limit:  0,
		Offset: 0,
		Total:  total,
	}

	if pagination.Limit != nil {
		resp.Limit = *pagination.Limit
	}
	if pagination.Offset != nil {
		resp.Offset = *pagination.Offset
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		exception.WriteApiError(
			w,
			exception.InternalServerError(err.Error()),
		)
	}
}
