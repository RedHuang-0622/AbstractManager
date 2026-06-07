package util

import (
	"context"
	"os"
	"testing"
	"time"
)

// =============================================================================
// GetDefaultDBTimeout
// =============================================================================

func TestGetDefaultDBTimeout_Default(t *testing.T) {
	os.Unsetenv("DB_TIMEOUT_SECONDS")
	got := GetDefaultDBTimeout()
	if got != 30*time.Second {
		t.Errorf("expected 30s default, got %v", got)
	}
}

func TestGetDefaultDBTimeout_EnvOverride(t *testing.T) {
	os.Setenv("DB_TIMEOUT_SECONDS", "15")
	defer os.Unsetenv("DB_TIMEOUT_SECONDS")

	got := GetDefaultDBTimeout()
	if got != 15*time.Second {
		t.Errorf("expected 15s from env, got %v", got)
	}
}

func TestGetDefaultDBTimeout_InvalidEnv(t *testing.T) {
	os.Setenv("DB_TIMEOUT_SECONDS", "not_a_number")
	defer os.Unsetenv("DB_TIMEOUT_SECONDS")

	got := GetDefaultDBTimeout()
	if got != 30*time.Second {
		t.Errorf("expected 30s fallback on invalid env, got %v", got)
	}
}

func TestGetDefaultDBTimeout_ZeroEnv(t *testing.T) {
	os.Setenv("DB_TIMEOUT_SECONDS", "0")
	defer os.Unsetenv("DB_TIMEOUT_SECONDS")

	got := GetDefaultDBTimeout()
	if got != 30*time.Second {
		t.Errorf("expected 30s fallback on zero value, got %v", got)
	}
}

func TestGetDefaultDBTimeout_NegativeEnv(t *testing.T) {
	os.Setenv("DB_TIMEOUT_SECONDS", "-5")
	defer os.Unsetenv("DB_TIMEOUT_SECONDS")

	got := GetDefaultDBTimeout()
	if got != 30*time.Second {
		t.Errorf("expected 30s fallback on negative value, got %v", got)
	}
}

// =============================================================================
// GetDefaultRedisTimeout
// =============================================================================

func TestGetDefaultRedisTimeout_Default(t *testing.T) {
	os.Unsetenv("REDIS_TIMEOUT_SECONDS")
	got := GetDefaultRedisTimeout()
	if got != 10*time.Second {
		t.Errorf("expected 10s default, got %v", got)
	}
}

func TestGetDefaultRedisTimeout_EnvOverride(t *testing.T) {
	os.Setenv("REDIS_TIMEOUT_SECONDS", "5")
	defer os.Unsetenv("REDIS_TIMEOUT_SECONDS")

	got := GetDefaultRedisTimeout()
	if got != 5*time.Second {
		t.Errorf("expected 5s from env, got %v", got)
	}
}

func TestGetDefaultRedisTimeout_InvalidEnv(t *testing.T) {
	os.Setenv("REDIS_TIMEOUT_SECONDS", "abc")
	defer os.Unsetenv("REDIS_TIMEOUT_SECONDS")

	got := GetDefaultRedisTimeout()
	if got != 10*time.Second {
		t.Errorf("expected 10s fallback on invalid env, got %v", got)
	}
}

func TestGetDefaultRedisTimeout_ZeroEnv(t *testing.T) {
	os.Setenv("REDIS_TIMEOUT_SECONDS", "0")
	defer os.Unsetenv("REDIS_TIMEOUT_SECONDS")

	got := GetDefaultRedisTimeout()
	if got != 10*time.Second {
		t.Errorf("expected 10s fallback on zero value, got %v", got)
	}
}

// =============================================================================
// GetDefaultDDLTimeout
// =============================================================================

func TestGetDefaultDDLTimeout_Default(t *testing.T) {
	os.Unsetenv("DDL_TIMEOUT_SECONDS")
	got := GetDefaultDDLTimeout()
	if got != 60*time.Second {
		t.Errorf("expected 60s default, got %v", got)
	}
}

func TestGetDefaultDDLTimeout_EnvOverride(t *testing.T) {
	os.Setenv("DDL_TIMEOUT_SECONDS", "120")
	defer os.Unsetenv("DDL_TIMEOUT_SECONDS")

	got := GetDefaultDDLTimeout()
	if got != 120*time.Second {
		t.Errorf("expected 120s from env, got %v", got)
	}
}

func TestGetDefaultDDLTimeout_InvalidEnv(t *testing.T) {
	os.Setenv("DDL_TIMEOUT_SECONDS", "xyz")
	defer os.Unsetenv("DDL_TIMEOUT_SECONDS")

	got := GetDefaultDDLTimeout()
	if got != 60*time.Second {
		t.Errorf("expected 60s fallback on invalid env, got %v", got)
	}
}

func TestGetDefaultDDLTimeout_ZeroEnv(t *testing.T) {
	os.Setenv("DDL_TIMEOUT_SECONDS", "0")
	defer os.Unsetenv("DDL_TIMEOUT_SECONDS")

	got := GetDefaultDDLTimeout()
	if got != 60*time.Second {
		t.Errorf("expected 60s fallback on zero value, got %v", got)
	}
}

// =============================================================================
// EnsureTimeout
// =============================================================================

func TestEnsureTimeout_NoDeadline(t *testing.T) {
	ctx := context.Background()
	derived, cancel := EnsureTimeout(ctx, 5*time.Second)
	defer cancel()

	deadline, ok := derived.Deadline()
	if !ok {
		t.Fatal("expected a deadline to be set")
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		t.Fatal("deadline is already in the past")
	}
	if remaining > 5*time.Second {
		t.Errorf("deadline too far in the future: %v", remaining)
	}
}

func TestEnsureTimeout_HasDeadline_KeepsOriginal(t *testing.T) {
	originalDeadline := time.Now().Add(2 * time.Second)
	ctx, origCancel := context.WithDeadline(context.Background(), originalDeadline)
	defer origCancel()

	derived, cancel := EnsureTimeout(ctx, 30*time.Second)
	defer cancel()

	deadline, ok := derived.Deadline()
	if !ok {
		t.Fatal("expected a deadline to be present")
	}

	diff := deadline.Sub(originalDeadline)
	if diff < -time.Millisecond || diff > time.Millisecond {
		t.Errorf("deadline changed: original=%v, got=%v, diff=%v", originalDeadline, deadline, diff)
	}
}

func TestEnsureTimeout_CancelFunc_NoDeadline(t *testing.T) {
	ctx := context.Background()
	_, cancel := EnsureTimeout(ctx, 1*time.Second)

	if cancel == nil {
		t.Fatal("cancel func is nil for ctx without deadline")
	}

	// Calling cancel should not panic
	cancel()
	// context.WithTimeout cancel is idempotent
	cancel()
}

func TestEnsureTimeout_CancelFunc_HasDeadline(t *testing.T) {
	ctx, origCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer origCancel()

	_, cancel := EnsureTimeout(ctx, 1*time.Second)

	if cancel == nil {
		t.Fatal("cancel func is nil")
	}

	// The returned cancel should be a no-op that doesn't cancel the parent
	cancel()
	select {
	case <-ctx.Done():
		t.Fatal("parent context was unexpectedly cancelled")
	default:
		// expected
	}
}

func TestEnsureTimeout_NoDeadline_CancelPropagates(t *testing.T) {
	ctx := context.Background()
	derived, cancel := EnsureTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	cancel()

	select {
	case <-derived.Done():
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("derived context was not cancelled")
	}
}

func TestEnsureTimeout_Concurrent(t *testing.T) {
	ctx := context.Background()
	done := make(chan struct{})

	for i := 0; i < 50; i++ {
		go func() {
			_, cancel := EnsureTimeout(ctx, 5*time.Second)
			cancel()
			done <- struct{}{}
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}
