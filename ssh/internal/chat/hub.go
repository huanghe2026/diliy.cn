// Package chat 实现 geoyuan SSH 聊天室的消息中心。
//
// 设计：
//   - Hub 是全局单例，每个已通过用户的 SSH 会话注册到 Hub
//   - 每个会话有一个缓冲 channel 接收广播消息，TUI 在独立 goroutine 里消费
//   - 消息类型：聊天消息 / 系统通知 / 私有消息
//   - 内置最近 200 条历史消息环形缓冲区，新用户进群会先收到历史
package chat

import (
	"fmt"
	"sync"
	"time"
)

// MessageType 消息类型
type MessageType int

const (
	MsgChat    MessageType = iota // 用户聊天消息（广播）
	MsgSystem                     // 系统通知（加入/离开）
	MsgPrivate                    // 私有消息（DM）
)

// Message 一条聊天室消息
type Message struct {
	Type       MessageType
	UserID     int64
	User       string
	TargetID   int64  // 私有消息目标 UserID（0 表示公开）
	TargetName string // 私有消息目标用户名
	Content    string
	Time       time.Time
}

// String 格式化消息为显示文本
func (m Message) String() string {
	ts := m.Time.Format("15:04:05")
	switch m.Type {
	case MsgSystem:
		return fmt.Sprintf("\x1b[90m[%s] *** %s ***\x1b[0m", ts, m.Content)
	case MsgPrivate:
		return fmt.Sprintf("\x1b[35m[%s] [DM] %s → %s:\x1b[0m %s", ts, m.User, m.TargetName, m.Content)
	default:
		return fmt.Sprintf("\x1b[36m[%s] %s:\x1b[0m %s", ts, m.User, m.Content)
	}
}

// Subscription 某个会话在 Hub 中的订阅
type Subscription struct {
	UserID int64
	User   string
	C      chan Message // 消息推送 channel，缓冲区 64
	hub    *Hub
}

const (
	historySize = 200 // 历史缓冲区大小（最近 200 条）
	channelSize = 64  // 订阅者 channel 缓冲区
)

// Hub 聊天室消息中心（全局单例）
type Hub struct {
	mu      sync.RWMutex
	subs    map[*Subscription]struct{}
	history []Message // 环形缓冲区（简单 append + 截断）
}

// NewHub 创建一个新的 Hub
func NewHub() *Hub {
	return &Hub{
		subs:    make(map[*Subscription]struct{}),
		history: make([]Message, 0, historySize),
	}
}

// Join 注册一个新用户，返回 Subscription。
// 自动把最近历史灌给新用户。
func (h *Hub) Join(userID int64, username string) *Subscription {
	sub := &Subscription{
		UserID: userID,
		User:   username,
		C:      make(chan Message, channelSize),
		hub:    h,
	}

	h.mu.Lock()
	h.subs[sub] = struct{}{}
	// 复制历史给新用户（channel 有缓冲，不会阻塞）
	history := make([]Message, len(h.history))
	copy(history, h.history)
	h.mu.Unlock()

	// 把历史塞进 channel（非阻塞，channel 满了跳过）
	for _, m := range history {
		select {
		case sub.C <- m:
		default:
			// channel 满了，丢这条历史，继续下一条
		}
	}

	// 广播系统通知
	h.Broadcast(Message{
		Type:    MsgSystem,
		User:    "system",
		Content: username + " 加入了聊天室",
		Time:    time.Now(),
	})
	return sub
}

// Leave 注销订阅，关闭 channel，广播离开通知
func (h *Hub) Leave(sub *Subscription) {
	if sub == nil {
		return
	}
	h.mu.Lock()
	if _, ok := h.subs[sub]; ok {
		delete(h.subs, sub)
		close(sub.C)
	}
	h.mu.Unlock()

	h.Broadcast(Message{
		Type:    MsgSystem,
		User:    "system",
		Content: sub.User + " 离开了聊天室",
		Time:    time.Now(),
	})
}

// SendChat 发送一条广播聊天消息（所有人能收到）
func (h *Hub) SendChat(userID int64, username, content string) {
	if content == "" {
		return
	}
	h.Broadcast(Message{
		Type:    MsgChat,
		UserID:  userID,
		User:    username,
		Content: content,
		Time:    time.Now(),
	})
}

// SendPrivate 发送一条私有消息（只发给目标用户；若目标不在线只存历史，双方在线时互发）
func (h *Hub) SendPrivate(fromID int64, fromName, toName, content string) {
	if content == "" {
		return
	}
	h.mu.RLock()
	// 找到目标订阅者
	var targetSub *Subscription
	for s := range h.subs {
		if s.User == toName {
			targetSub = s
			break
		}
	}
	h.mu.RUnlock()

	msg := Message{
		Type:       MsgPrivate,
		UserID:     fromID,
		User:       fromName,
		TargetName: toName,
		Content:    content,
		Time:       time.Now(),
	}

	// 存历史
	h.addHistory(msg)

	if targetSub == nil {
		// 目标不在线：通知发送者"对方不在线，消息已存历史"
		h.addHistory(Message{
			Type:       MsgSystem,
			User:       "system",
			Content:    fmt.Sprintf("%s 当前不在线，消息已暂存", toName),
			Time:       time.Now(),
			TargetName: fromName, // 只推给发送者
		})
		return
	}

	// 推给目标
	h.pushTo(targetSub, msg)

	// 推给自己（让发送者确认已送达）
	h.mu.RLock()
	var selfSub *Subscription
	for s := range h.subs {
		if s.UserID == fromID {
			selfSub = s
			break
		}
	}
	h.mu.RUnlock()
	if selfSub != nil && selfSub != targetSub {
		h.pushTo(selfSub, msg)
	}
}

// Broadcast 向所有订阅者推送消息（内部方法），同时写入历史
func (h *Hub) Broadcast(msg Message) {
	h.addHistory(msg)
	h.doBroadcast(msg)
}

// doBroadcast 纯广播（不存历史），供 SendPrivate 内部使用
func (h *Hub) doBroadcast(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.subs {
		h.pushTo(sub, msg)
	}
}

// pushTo 非阻塞推送单条消息给订阅者
func (h *Hub) pushTo(sub *Subscription, msg Message) {
	if sub == nil {
		return
	}
	select {
	case sub.C <- msg:
	default:
		// channel 满了跳过，慢消费者丢消息但不拖垮 Hub
	}
}

// addHistory 追加消息到历史缓冲区（超了就截断）
func (h *Hub) addHistory(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.history) >= historySize {
		// 截断最旧的 1/4，避免频繁 slice 分配
		h.history = h.history[historySize/4:]
	}
	h.history = append(h.history, msg)
}

// Count 返回当前在线人数
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}

// Users 返回当前在线用户名列表
func (h *Hub) Users() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	names := make([]string, 0, len(h.subs))
	for sub := range h.subs {
		names = append(names, sub.User)
	}
	return names
}

// FindUser 按用户名查在线订阅者（返回 userID，找不到返回 0）
func (h *Hub) FindUser(name string) int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.subs {
		if sub.User == name {
			return sub.UserID
		}
	}
	return 0
}
