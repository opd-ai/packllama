package ui

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConversationManager maintains a set of conversations stored in a directory.
type ConversationManager struct {
	dir           string
	conversations []*Conversation
	active        *Conversation
}

// NewConversationManager creates a manager rooted at dir.
// Call Load to populate it from existing files.
func NewConversationManager(dir string) *ConversationManager {
	return &ConversationManager{dir: dir}
}

// Load reads all *.json files from the manager's directory.
// Corrupt files are silently skipped.
func (m *ConversationManager) Load() error {
	entries, err := filepath.Glob(filepath.Join(m.dir, "*.json"))
	if err != nil {
		return fmt.Errorf("glob conversations: %w", err)
	}
	m.conversations = m.conversations[:0]
	for _, e := range entries {
		c, err := LoadConversation(e)
		if err != nil {
			continue
		}
		m.conversations = append(m.conversations, c)
	}
	return nil
}

// New creates, registers, and activates a fresh conversation.
func (m *ConversationManager) New(title string) *Conversation {
	c := NewConversation(title)
	m.conversations = append(m.conversations, c)
	m.active = c
	return c
}

// Active returns the currently active conversation, or nil.
func (m *ConversationManager) Active() *Conversation { return m.active }

// SetActive switches the active conversation to the one at index i.
func (m *ConversationManager) SetActive(i int) {
	if i >= 0 && i < len(m.conversations) {
		m.active = m.conversations[i]
	}
}

// List returns all managed conversations.
func (m *ConversationManager) List() []*Conversation { return m.conversations }

// SaveActive persists the active conversation to disk. No-ops when nil.
func (m *ConversationManager) SaveActive() error {
	if m.active == nil {
		return nil
	}
	return m.active.Save(m.dir)
}

// Delete removes the conversation at index i and erases its file.
func (m *ConversationManager) Delete(i int) error {
	if i < 0 || i >= len(m.conversations) {
		return fmt.Errorf("index %d out of range", i)
	}
	c := m.conversations[i]
	path := filepath.Join(m.dir, c.ID+".json")
	_ = os.Remove(path) // best-effort
	m.conversations = append(m.conversations[:i], m.conversations[i+1:]...)
	if m.active == c {
		m.active = nil
	}
	return nil
}
