package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketService WebSocket服务
type WebSocketService struct {
	connections map[uuid.UUID]*websocket.Conn
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		connections: make(map[uuid.UUID]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有来源（生产环境中应该更严格）
				return true
			},
		},
	}
}

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// HandleWebSocket 处理WebSocket连接
func (s *WebSocketService) HandleWebSocket(c *gin.Context) {
	// 获取用户ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 升级HTTP连接为WebSocket连接
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 添加连接到管理器
	s.addConnection(userID, conn)
	defer s.removeConnection(userID)

	log.Printf("用户 %s 建立WebSocket连接", userID)

	// 发送欢迎消息
	welcomeMsg := NotificationMessage{
		Type:      "system",
		Title:     "连接成功",
		Content:   "实时通知已启用",
		Timestamp: getCurrentTimestamp(),
	}
	s.sendToUser(userID, welcomeMsg)

	// 监听消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}

		// 处理收到的消息（心跳包等）
		s.handleMessage(userID, message)
	}
}

// addConnection 添加连接
func (s *WebSocketService) addConnection(userID uuid.UUID, conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果用户已有连接，关闭旧连接
	if oldConn, exists := s.connections[userID]; exists {
		oldConn.Close()
	}

	s.connections[userID] = conn
}

// removeConnection 移除连接
func (s *WebSocketService) removeConnection(userID uuid.UUID) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if conn, exists := s.connections[userID]; exists {
		conn.Close()
		delete(s.connections, userID)
		log.Printf("用户 %s 断开WebSocket连接", userID)
	}
}

// sendToUser 发送消息给指定用户
func (s *WebSocketService) sendToUser(userID uuid.UUID, message NotificationMessage) error {
	s.mutex.RLock()
	conn, exists := s.connections[userID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("用户 %s 未连接", userID)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		// 连接可能已断开，移除连接
		s.removeConnection(userID)
		return fmt.Errorf("发送消息失败: %v", err)
	}

	return nil
}

// BroadcastToUsers 广播消息给多个用户
func (s *WebSocketService) BroadcastToUsers(userIDs []uuid.UUID, message NotificationMessage) {
	for _, userID := range userIDs {
		if err := s.sendToUser(userID, message); err != nil {
			log.Printf("发送消息给用户 %s 失败: %v", userID, err)
		}
	}
}

// BroadcastToAll 广播消息给所有连接的用户
func (s *WebSocketService) BroadcastToAll(message NotificationMessage) {
	s.mutex.RLock()
	userIDs := make([]uuid.UUID, 0, len(s.connections))
	for userID := range s.connections {
		userIDs = append(userIDs, userID)
	}
	s.mutex.RUnlock()

	s.BroadcastToUsers(userIDs, message)
}

// GetConnectedUsers 获取已连接的用户列表
func (s *WebSocketService) GetConnectedUsers() []uuid.UUID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	users := make([]uuid.UUID, 0, len(s.connections))
	for userID := range s.connections {
		users = append(users, userID)
	}
	return users
}

// IsUserConnected 检查用户是否在线
func (s *WebSocketService) IsUserConnected(userID uuid.UUID) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.connections[userID]
	return exists
}

// handleMessage 处理收到的消息
func (s *WebSocketService) handleMessage(userID uuid.UUID, message []byte) {
	// 解析消息
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("解析消息失败: %v", err)
		return
	}

	// 处理心跳包
	if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
		pongMsg := NotificationMessage{
			Type:      "pong",
			Timestamp: getCurrentTimestamp(),
		}
		s.sendToUser(userID, pongMsg)
	}
}

// SendNotification 发送通知（对外接口）
func (s *WebSocketService) SendNotification(userID uuid.UUID, notificationType, title, content string, data interface{}) error {
	message := NotificationMessage{
		Type:      notificationType,
		Title:     title,
		Content:   content,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
	}

	return s.sendToUser(userID, message)
}

// SendSystemNotification 发送系统通知
func (s *WebSocketService) SendSystemNotification(userID uuid.UUID, title, content string) error {
	return s.SendNotification(userID, "system", title, content, nil)
}

// SendProjectNotification 发送项目通知
func (s *WebSocketService) SendProjectNotification(userID uuid.UUID, projectID uuid.UUID, title, content string) error {
	data := map[string]interface{}{
		"project_id": projectID.String(),
	}
	return s.SendNotification(userID, "project", title, content, data)
}

// SendHomeworkNotification 发送作业通知
func (s *WebSocketService) SendHomeworkNotification(userID uuid.UUID, homeworkID uuid.UUID, title, content string) error {
	data := map[string]interface{}{
		"homework_id": homeworkID.String(),
	}
	return s.SendNotification(userID, "homework", title, content, data)
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
