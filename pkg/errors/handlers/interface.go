package handlers

import (
	"context"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
)

// ErrorHandler defines the interface for handling specific error categories
type ErrorHandler interface {
	Handle(ctx context.Context, err *types.AppError) *types.AppError
	CanHandle(err *types.AppError) bool
	GetCategory() types.ErrorCategory
}

// BaseHandler provides common functionality for error handlers
type BaseHandler struct {
	category types.ErrorCategory
}

// GetCategory returns the error category this handler supports
func (h *BaseHandler) GetCategory() types.ErrorCategory {
	return h.category
}

// CanHandle checks if this handler can handle the given error
func (h *BaseHandler) CanHandle(err *types.AppError) bool {
	return types.GetErrorCategory(err.Code) == h.category
}