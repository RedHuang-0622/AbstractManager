package filter_translator

import (
	"math"
	"testing"
)

// =============================================================================
// UT-008 ~ UT-013: toFloat64 白盒测试（函数未导出，需在包内测试）
// =============================================================================

func TestToFloat64_Int(t *testing.T) {
	result, err := toFloat64(123)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 123.0 {
		t.Errorf("expected 123.0, got %v", result)
	}
}

func TestToFloat64_Int64(t *testing.T) {
	result, err := toFloat64(int64(99))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 99.0 {
		t.Errorf("expected 99.0, got %v", result)
	}
}

func TestToFloat64_Float64(t *testing.T) {
	result, err := toFloat64(45.67)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if math.Abs(result-45.67) > 0.0001 {
		t.Errorf("expected 45.67, got %v", result)
	}
}

func TestToFloat64_StringValid(t *testing.T) {
	result, err := toFloat64("45.67")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if math.Abs(result-45.67) > 0.0001 {
		t.Errorf("expected 45.67, got %v", result)
	}
}

func TestToFloat64_StringInteger(t *testing.T) {
	result, err := toFloat64("42")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 42.0 {
		t.Errorf("expected 42.0, got %v", result)
	}
}

func TestToFloat64_StringInvalid(t *testing.T) {
	_, err := toFloat64("abc")
	if err == nil {
		t.Fatal("expected error for invalid string")
	}
}

func TestToFloat64_Bool(t *testing.T) {
	_, err := toFloat64(true)
	if err == nil {
		t.Fatal("expected error for bool")
	}
}

func TestToFloat64_Nil(t *testing.T) {
	_, err := toFloat64(nil)
	if err == nil {
		t.Fatal("expected error for nil")
	}
}

func TestToFloat64_Struct(t *testing.T) {
	_, err := toFloat64(struct{}{})
	if err == nil {
		t.Fatal("expected error for struct")
	}
}

func TestToFloat64_Float32(t *testing.T) {
	// float32 is not handled by toFloat64's type switch → "unknown type"
	_, err := toFloat64(float32(3.14))
	if err == nil {
		t.Fatal("expected error for float32 (not in type switch)")
	}
}

func TestToFloat64_NegativeInt(t *testing.T) {
	result, err := toFloat64(-10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != -10.0 {
		t.Errorf("expected -10.0, got %v", result)
	}
}

func TestToFloat64_ZeroInt(t *testing.T) {
	result, err := toFloat64(0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 0.0 {
		t.Errorf("expected 0.0, got %v", result)
	}
}
