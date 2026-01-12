package core

import (
	"sync"
	"time"
)

type Step string

const (
	StepAwaitingContainerName Step = "awaiting_container_name"
	StepAwaitingConfirm       Step = "awaiting_confirm"
)

type Action string

const (
	ActionRestart     Action = "restart"
	ActionStop        Action = "stop"
	ActionForceUpdate Action = "force_update"
)

func ActionFromEventKey(key string) Action {
	switch key {
	case "unraid.action.restart":
		return ActionRestart
	case "unraid.action.stop":
		return ActionStop
	case "unraid.action.force_update":
		return ActionForceUpdate
	default:
		return ""
	}
}

func (a Action) DisplayName() string {
	switch a {
	case ActionRestart:
		return "重启"
	case ActionStop:
		return "停止"
	case ActionForceUpdate:
		return "强制更新"
	default:
		return "未知动作"
	}
}

type ConversationState struct {
	Step          Step
	Action        Action
	ContainerName string
	ExpiresAt     time.Time
}

type StateStore struct {
	ttl  time.Duration
	mu   sync.Mutex
	data map[string]ConversationState
}

func NewStateStore(ttl time.Duration) *StateStore {
	return &StateStore{
		ttl:  ttl,
		data: make(map[string]ConversationState),
	}
}

func (s *StateStore) Get(userID string) (ConversationState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.data[userID]
	if !ok {
		return ConversationState{}, false
	}
	if time.Now().After(state.ExpiresAt) {
		delete(s.data, userID)
		return ConversationState{}, false
	}
	return state, true
}

func (s *StateStore) Set(userID string, state ConversationState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state.ExpiresAt = time.Now().Add(s.ttl)
	s.data[userID] = state
}

func (s *StateStore) Clear(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, userID)
}
