package middleware

import (
	"context"
	"testing"
)

// This test demonstrates the behavior of unexported context keys

func TestUnexportedContextKey(t *testing.T) {
	ctx := context.Background()

	// Test: Using the type-safe helper function
	ctx = WithUserID(ctx, "test-user-id")
	userID := GetUserIDFromContext(ctx)

	if userID != "test-user-id" {
		t.Errorf("Expected 'test-user-id', got %v", userID)
	}

	// Test: String literal won't retrieve the value (different type)
	val := ctx.Value("user_id")
	if val != nil {
		t.Errorf("String literal should return nil (different type), got %v", val)
	}

	// Test: Empty context returns empty string
	emptyCtx := context.Background()
	emptyUserID := GetUserIDFromContext(emptyCtx)
	if emptyUserID != "" {
		t.Errorf("Expected empty string for empty context, got %v", emptyUserID)
	}
}
