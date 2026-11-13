package handler

import (
	"errors"
	"net/http"
	"strconv"

	"be-presensi/internal/modules/activity_comment/dto"
	"be-presensi/internal/modules/activity_comment/entity"
	"be-presensi/internal/modules/activity_comment/service"
	"be-presensi/internal/modules/auth/middleware"
	"be-presensi/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ActivityCommentHandler struct {
	service service.ActivityCommentService
}

func NewActivityCommentHandler(
	activityCommentService service.ActivityCommentService,
) *ActivityCommentHandler {
	return &ActivityCommentHandler{
		service: activityCommentService,
	}
}

func (h *ActivityCommentHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(
			c,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	var req dto.CreateActivityCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid activity comment request",
		)
		return
	}

	comment, err := h.service.Create(
		c.Request.Context(),
		userID,
		req,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(
		c,
		http.StatusCreated,
		"activity comment created successfully",
		toActivityCommentResponse(comment),
	)
}

func toActivityCommentResponse(
	comment *entity.ActivityComment,
) dto.ActivityCommentResponse {
	return dto.ActivityCommentResponse{
		ID:              comment.ID,
		DailyActivityID: comment.DailyActivityID,
		KepsekID:        comment.KepsekID,
		Comment:         comment.Comment,
		WANotified:      comment.WANotified,
		CreatedAt:       comment.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (h *ActivityCommentHandler) handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrActivityCommentForbidden):
		response.Error(
			c,
			http.StatusForbidden,
			err.Error(),
		)

	case errors.Is(err, service.ErrDailyActivityNotFound),
		errors.Is(err, service.ErrCommentOwnerNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(
			c,
			http.StatusNotFound,
			err.Error(),
		)

	case errors.Is(err, service.ErrActivityCommentRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

	default:
		c.Error(err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"failed to create activity comment",
		)
	}
}

func (h *ActivityCommentHandler) GetByDailyActivityID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID <= 0 {
		response.Error(
			c,
			http.StatusUnauthorized,
			"invalid authenticated user",
		)
		return
	}

	dailyActivityID, err := strconv.ParseInt(
		c.Param("id"),
		10,
		64,
	)
	if err != nil || dailyActivityID <= 0 {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid daily activity id",
		)
		return
	}

	comments, err := h.service.GetByDailyActivityID(
		c.Request.Context(),
		userID,
		dailyActivityID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	responses := make([]dto.ActivityCommentResponse, 0, len(comments))

	for _, comment := range comments {
		responses = append(
			responses,
			toActivityCommentResponse(comment),
		)
	}

	response.Success(
		c,
		http.StatusOK,
		"activity comments retrieved successfully",
		responses,
	)
}
