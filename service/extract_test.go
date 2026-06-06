package service

import (
	"encoding/json"
	"testing"
)

// =============================================================================
// UT-304 ~ UT-309: extractID / extractIDFromKey 白盒测试
// =============================================================================

type extractTestUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Age      int    `json:"age"`
}

type extractNoIDUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type extractFloatIDUser struct {
	ID    float64 `json:"id"`
	Value string  `json:"value"`
}

func TestExtractID_Valid(t *testing.T) {
	user := extractTestUser{ID: 42, Username: "alice", Age: 30}
	id, ok := extractID(&user)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if id != 42 {
		t.Errorf("expected id=42, got %d", id)
	}
}

func TestExtractID_ZeroID(t *testing.T) {
	user := extractTestUser{ID: 0, Username: "new_user"}
	id, ok := extractID(&user)
	if !ok {
		t.Fatal("expected ok=true (0 is valid ID)")
	}
	if id != 0 {
		t.Errorf("expected id=0, got %d", id)
	}
}

func TestExtractID_NoIDField(t *testing.T) {
	user := extractNoIDUser{Name: "bob", Email: "bob@test.com"}
	_, ok := extractID(&user)
	if ok {
		t.Fatal("expected ok=false for struct without id field")
	}
}

func TestExtractID_FloatID(t *testing.T) {
	// JSON numbers unmarshal as float64, so float64 id field should work
	user := extractFloatIDUser{ID: 99.0, Value: "test"}
	id, ok := extractID(&user)
	if !ok {
		t.Fatal("expected ok=true for float64 id")
	}
	if id != 99 {
		t.Errorf("expected id=99, got %d", id)
	}
}

func TestExtractID_NilInput(t *testing.T) {
	// Passing nil pointer — marshal should fail gracefully
	var user *extractTestUser = nil
	_, ok := extractID(user)
	if ok {
		t.Fatal("expected ok=false for nil input")
	}
}

func TestExtractID_StringID(t *testing.T) {
	// ID field as string should NOT be converted (not a float64 in JSON)
	type stringIDUser struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	user := stringIDUser{ID: "uuid-123", Name: "test"}
	_, ok := extractID(&user)
	if ok {
		t.Fatal("expected ok=false for string id")
	}
}

func TestExtractIDFromKey_Valid(t *testing.T) {
	id, err := extractIDFromKey("user:123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 123 {
		t.Errorf("expected id=123, got %d", id)
	}
}

func TestExtractIDFromKey_ZeroID(t *testing.T) {
	id, err := extractIDFromKey("user:0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 0 {
		t.Errorf("expected id=0, got %d", id)
	}
}

func TestExtractIDFromKey_LargeID(t *testing.T) {
	id, err := extractIDFromKey("user:4294967295")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 4294967295 {
		t.Errorf("expected id=4294967295, got %d", id)
	}
}

func TestExtractIDFromKey_NoColon(t *testing.T) {
	_, err := extractIDFromKey("simplekey")
	if err == nil {
		t.Fatal("expected error for key without colon")
	}
}

func TestExtractIDFromKey_NonNumeric(t *testing.T) {
	_, err := extractIDFromKey("user:abc")
	if err == nil {
		t.Fatal("expected error for non-numeric ID")
	}
}

func TestExtractIDFromKey_TripleColon(t *testing.T) {
	id, err := extractIDFromKey("cache:user:456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 456 {
		t.Errorf("expected id=456, got %d", id)
	}
}

func TestExtractIDFromKey_FourParts(t *testing.T) {
	id, err := extractIDFromKey("a:b:c:999")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 999 {
		t.Errorf("expected id=999, got %d", id)
	}
}

func TestExtractIDFromKey_EmptyString(t *testing.T) {
	_, err := extractIDFromKey("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestExtractIDFromKey_ColonOnly(t *testing.T) {
	_, err := extractIDFromKey(":")
	if err == nil {
		t.Fatal("expected error for colon-only")
	}
}

// =============================================================================
// Round-trip: json marshal → extractID → buildCacheKey → extractIDFromKey
// =============================================================================

func TestExtractID_RoundTrip(t *testing.T) {
	user := extractTestUser{ID: 777, Username: "roundtrip"}

	// 1. marshal to JSON (simulating what happens in WritedownSingle)
	data, err := json.Marshal(&user)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// 2. unmarshal back and extract ID
	var tempMap map[string]interface{}
	if err := json.Unmarshal(data, &tempMap); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	idFloat, ok := tempMap["id"].(float64)
	if !ok {
		t.Fatal("id not found or not float64 in unmarshalled map")
	}
	id := uint(idFloat)
	if id != 777 {
		t.Errorf("expected 777, got %d", id)
	}

	// 3. build key
	sm := NewServiceManager(extractTestUser{})
	sm.CacheKeyName = "extract_test_user"
	key := sm.buildCacheKey(id)

	// 4. extract from key
	extracted, err := extractIDFromKey(key)
	if err != nil {
		t.Fatalf("extractIDFromKey failed: %v", err)
	}
	if extracted != 777 {
		t.Errorf("round-trip failed: expected 777, got %d", extracted)
	}
}

func TestBuildCacheKey_DefaultFormat(t *testing.T) {
	sm := NewServiceManager(extractTestUser{})
	sm.CacheKeyName = "extract_test_user"

	key := sm.buildCacheKey(uint(42))
	if key != "extract_test_user:42" {
		t.Errorf("expected 'extract_test_user:42', got '%s'", key)
	}
}

func TestBuildCacheKey_WithType(t *testing.T) {
	sm := NewServiceManager(extractTestUser{})
	sm.CacheKeyType = "cache"
	sm.CacheKeyName = "extract_test_user"

	key := sm.buildCacheKey(uint(42))
	if key != "cache:extract_test_user:42" {
		t.Errorf("expected 'cache:extract_test_user:42', got '%s'", key)
	}
}

func TestBuildCacheKey_StringID(t *testing.T) {
	sm := NewServiceManager(extractTestUser{})
	sm.CacheKeyName = "extract_test_user"

	key := sm.buildCacheKey("uuid-abc-123")
	if key != "extract_test_user:uuid-abc-123" {
		t.Errorf("expected 'extract_test_user:uuid-abc-123', got '%s'", key)
	}
}

func TestExtractID_JSONNumberEdgeCase(t *testing.T) {
	// Large number: JSON number → float64 can lose precision for very large integers
	// This test documents the known limitation
	user := extractFloatIDUser{ID: float64(uint(1 << 50)), Value: "large"}
	id, ok := extractID(&user)
	if !ok {
		t.Fatal("expected ok=true")
	}
	// float64 has ~53 bits of mantissa, so 1<<50 should be precise
	if float64(id) != float64(uint(1<<50)) {
		t.Logf("note: float64 precision limitation — id=%d, expected=%d", id, uint(1<<50))
	}
}
