package handler

import (
	"net/http"

	"be-presensi/internal/modules/notification/service"

	"github.com/gin-gonic/gin"
)

type WhatsAppHandler struct {
	notificationService service.NotificationService
}

type testWhatsAppRequest struct {
	Target  string `json:"target" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func NewWhatsAppHandler(
	notificationService service.NotificationService,
) *WhatsAppHandler {
	return &WhatsAppHandler{
		notificationService: notificationService,
	}
}

func (h *WhatsAppHandler) TestSend(c *gin.Context) {
	var req testWhatsAppRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "target and message are required",
		})
		return
	}

	result, err := h.notificationService.SendWhatsAppMessage(
		c.Request.Context(),
		req.Target,
		req.Message,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "WhatsApp message sent successfully",
		"data":    result,
	})
}
