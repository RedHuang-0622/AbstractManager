// Package testutil provides shared test fixtures and helpers for AbstractManager tests.
package testutil

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// TestUser mirrors the example User model for use in tests.
type TestUser struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;size:64" json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Status   string `json:"status"`
}

// Product is a test model for multi-field key template tests.
type Product struct {
	ID       uint    `json:"id"`
	Category string  `json:"category"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
}

// NestedContainer holds a nested struct for nested-field extraction tests.
type NestedContainer struct {
	User struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

// SetupMiniRedis spins up an in-memory miniredis and returns a go-redis client.
func SetupMiniRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

// SeedRedis populates Redis with the given key→JSON-string map.
func SeedRedis(ctx context.Context, t *testing.T, client *redis.Client, data map[string]string) {
	t.Helper()
	for k, v := range data {
		require.NoError(t, client.Set(ctx, k, v, 0).Err())
	}
}
