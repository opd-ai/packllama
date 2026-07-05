package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConversationRole identifies who authored a message.
type ConversationRole string

const (
	// RoleUser is the human participant.
	RoleUser ConversationRole = "user"
	// RoleAssistant is the AI model.
	RoleAssistant ConversationRole = "assistant"
	// RoleSystem is a system-level prompt.
	RoleSystem ConversationRole = "system"
)

// ConversationMessage is a single turn in a conversation.
type ConversationMessage struct {
	Role    ConversationRole `json:"role"`
	Content string           `json:"content"`
}

// Conversation is the complete history for one chat session.
type Conversation struct {
	ID        string                `json:"id"`
	Title     string                `json:"title"`
	Messages  []ConversationMessage `json:"messages"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// NewConversation creates a Conversation with a generated ID.
func NewConversation(title string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        fmt.Sprintf("conv_%d", now.UnixNano()),
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage appends a new message and updates UpdatedAt.
func (c *Conversation) AddMessage(role ConversationRole, content string) {
	c.Messages = append(c.Messages, ConversationMessage{Role: role, Content: content})
	c.UpdatedAt = time.Now()
}

// EditMessage replaces the content of message at index i.
func (c *Conversation) EditMessage(i int, content string) error {
	if i < 0 || i >= len(c.Messages) {
		return fmt.Errorf("message index %d out of range", i)
	}
	c.Messages[i].Content = content
	c.UpdatedAt = time.Now()
	return nil
}

// DeleteMessage removes the message at index i.
func (c *Conversation) DeleteMessage(i int) error {
	if i < 0 || i >= len(c.Messages) {
		return fmt.Errorf("message index %d out of range", i)
	}
	c.Messages = append(c.Messages[:i], c.Messages[i+1:]...)
	c.UpdatedAt = time.Now()
	return nil
}

// Save writes the conversation as JSON to dir/<id>.json.
func (c *Conversation) Save(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create conversations dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal conversation: %w", err)
	}
	path := filepath.Join(dir, c.ID+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write conversation: %w", err)
	}
	return nil
}

// LoadConversation reads and unmarshals a conversation from path.
func LoadConversation(path string) (*Conversation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read conversation: %w", err)
	}
	var c Conversation
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshal conversation: %w", err)
	}
	return &c, nil
}
