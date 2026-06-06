package service

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// =============================================================================
// UT-310 ~ UT-311: applyQueryOptions 白盒测试
// =============================================================================

type queryOptTestModel struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Status string `json:"status"`
}

func setupOptTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open failed: %v", err)
	}
	return gormDB
}

func makeOptSM() *ServiceManager[queryOptTestModel] {
	return NewServiceManager(queryOptTestModel{})
}

// ---------------------------------------------------------------------------
// Nil / empty options
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_Nil(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, nil)
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_Empty(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_Pagination(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Page: 2, PageSize: 10,
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_PageZero(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	// Page 0 → treated as page 1 (offset 0)
	result := sm.applyQueryOptions(db, &QueryOptions{
		Page: 0, PageSize: 20,
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_PageSizeNegative(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	// PageSize <= 0 means no pagination applied
	result := sm.applyQueryOptions(db, &QueryOptions{
		Page: 1, PageSize: 0,
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// OrderBy
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_OrderBy_Asc(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		OrderBy: "age", Order: "ASC",
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_OrderBy_Desc(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		OrderBy: "created_at", Order: "DESC",
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_OrderBy_DefaultAsc(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	// Order="" defaults to ASC
	result := sm.applyQueryOptions(db, &QueryOptions{
		OrderBy: "id", Order: "",
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_Order_InvalidDirection(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		OrderBy: "id", Order: "DELETE",
	})
	// Should have added an error
	if result.Error == nil {
		t.Fatal("expected error for invalid order direction")
	}
}

func TestApplyQueryOptions_Order_CaseInsensitive(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	// "asc" lowercase → should work (uppercased to "ASC")
	result := sm.applyQueryOptions(db, &QueryOptions{
		OrderBy: "id", Order: "asc",
	})
	if result.Error != nil {
		t.Errorf("expected no error for lowercase 'asc', got %v", result.Error)
	}
}

func TestApplyQueryOptions_OrderBy_SQLInjection(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		OrderBy: "id;DROP TABLE users--", Order: "ASC",
	})
	if result.Error == nil {
		t.Fatal("expected error for SQL injection in OrderBy")
	}
}

// ---------------------------------------------------------------------------
// Group / Having
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_Group(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Group: "status",
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_Group_SQLInjection(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Group: "status;DELETE FROM users--",
	})
	if result.Error == nil {
		t.Fatal("expected error for SQL injection in Group")
	}
}

func TestApplyQueryOptions_Having_Legacy(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Having: map[string]interface{}{
			"count": 5,
		},
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_Having_LegacySQLInjection(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Having: map[string]interface{}{
			"count;DROP": 1,
		},
	})
	if result.Error == nil {
		t.Fatal("expected error for SQL injection in Having key")
	}
}

// ---------------------------------------------------------------------------
// HavingConditions (structured)
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_HavingConditions_Equal(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		HavingConditions: []HavingCondition{
			{Field: "total", Operator: "=", Value: 100},
		},
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_HavingConditions_AllOperators(t *testing.T) {
	sm := makeOptSM()

	validOps := []string{"=", "!=", ">", ">=", "<", "<="}
	for _, op := range validOps {
		t.Run("op_"+op, func(t *testing.T) {
			db := setupOptTestDB(t)
			result := sm.applyQueryOptions(db, &QueryOptions{
				HavingConditions: []HavingCondition{
					{Field: "cnt", Operator: op, Value: 10},
				},
			})
			if result.Error != nil {
				t.Errorf("operator %s: expected no error, got %v", op, result.Error)
			}
		})
	}
}

func TestApplyQueryOptions_HavingConditions_InvalidOperator(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		HavingConditions: []HavingCondition{
			{Field: "cnt", Operator: "LIKE", Value: 10},
		},
	})
	if result.Error == nil {
		t.Fatal("expected error for invalid HavingCondition operator")
	}
}

func TestApplyQueryOptions_HavingConditions_SQLInjectionOperator(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		HavingConditions: []HavingCondition{
			{Field: "cnt", Operator: "; DROP TABLE users--", Value: 1},
		},
	})
	if result.Error == nil {
		t.Fatal("expected error for SQL injection in HavingCondition operator")
	}
}

func TestApplyQueryOptions_HavingConditions_SQLInjectionField(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		HavingConditions: []HavingCondition{
			{Field: "cnt;DELETE FROM users", Operator: "=", Value: 1},
		},
	})
	if result.Error == nil {
		t.Fatal("expected error for SQL injection in HavingCondition field")
	}
}

// ---------------------------------------------------------------------------
// Select / Distinct / Preload
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_Select(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Select: []string{"id", "name"},
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_Distinct(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Distinct: true,
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyQueryOptions_Preload(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Preload: []string{"Orders", "Profile"},
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Combined options
// ---------------------------------------------------------------------------

func TestApplyQueryOptions_AllCombined(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyQueryOptions(db, &QueryOptions{
		Page:     1,
		PageSize: 20,
		OrderBy:  "age",
		Order:    "DESC",
		Select:   []string{"id", "name", "age"},
		Distinct: false,
		Group:    "status",
		HavingConditions: []HavingCondition{
			{Field: "cnt", Operator: ">", Value: 0},
		},
		Preload: []string{"Profile"},
	})
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// applyTableName
// ---------------------------------------------------------------------------

func TestApplyTableName_Default(t *testing.T) {
	sm := makeOptSM()
	db := setupOptTestDB(t)

	result := sm.applyTableName(db)
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyTableName_WithSchema(t *testing.T) {
	sm := makeOptSM()
	sm.Schema = "private"
	db := setupOptTestDB(t)

	result := sm.applyTableName(db)
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestApplyTableName_PublicSchema(t *testing.T) {
	sm := makeOptSM()
	sm.Schema = "public" // public is treated as no schema prefix
	db := setupOptTestDB(t)

	result := sm.applyTableName(db)
	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}
