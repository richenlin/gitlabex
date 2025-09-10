package handlers

import (
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	wsService *services.WebSocketService
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(wsService *services.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		wsService: wsService,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	h.wsService.HandleWebSocket(c)
}
