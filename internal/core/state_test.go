// StateStore 与动作属性单元测试。
package core

import (
	"testing"
	"time"
)

func TestStateStore_TTL(t *testing.T) {
	t.Parallel()

	store := NewStateStore(20 * time.Millisecond)
	t.Cleanup(store.Close)
	store.Set("u", ConversationState{
		Step:   StepAwaitingContainerName,
		Action: ActionUnraidRestart,
	})

	if _, ok := store.Get("u"); !ok {
		t.Fatalf("Get() ok = false, want true")
	}

	time.Sleep(25 * time.Millisecond)
	if _, ok := store.Get("u"); ok {
		t.Fatalf("Get() ok = true, want false")
	}
}

func TestStateStore_JanitorPrunesWithoutGet(t *testing.T) {
	t.Parallel()

	store := NewStateStore(20 * time.Millisecond)
	t.Cleanup(store.Close)

	store.Set("u", ConversationState{
		Step:   StepAwaitingContainerName,
		Action: ActionUnraidRestart,
	})

	time.Sleep(120 * time.Millisecond)

	store.mu.Lock()
	_, ok := store.data["u"]
	store.mu.Unlock()
	if ok {
		t.Fatalf("janitor did not prune expired state")
	}
}

func TestAction_RequiresConfirm(t *testing.T) {
	t.Parallel()

	if !ActionUnraidRestart.RequiresConfirm() {
		t.Fatalf("ActionUnraidRestart RequiresConfirm() = false, want true")
	}
	if ActionUnraidViewStatus.RequiresConfirm() {
		t.Fatalf("ActionUnraidViewStatus RequiresConfirm() = true, want false")
	}
}
