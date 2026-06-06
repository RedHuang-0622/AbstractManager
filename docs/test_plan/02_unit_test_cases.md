# 单元测试用例详细清单

> 覆盖所有函数/方法级别的独立测试。使用 mock/sqlmock/miniredis 隔离外部依赖。

## util/filter_translator

| ID | 测试函数 | 输入 | 预期输出 | 优先级 |
|----|---------|------|---------|--------|
| UT-001 | `TestValidateSQLIdentifier_Valid` | `"user_name"` | nil error | P0 |
| UT-002 | `TestValidateSQLIdentifier_StartsWithNumber` | `"1user"` | error | P0 |
| UT-003 | `TestValidateSQLIdentifier_SQLInjection_Drop` | `"id;DROP TABLE"` | error | P0 |
| UT-004 | `TestValidateSQLIdentifier_SQLInjection_Comment` | `"id--"` | error | P0 |
| UT-005 | `TestValidateSQLIdentifier_Empty` | `""` | error | P1 |
| UT-006 | `TestValidateSQLIdentifier_SpecialChars` | `"id@x"` | error | P1 |
| UT-007 | `TestValidateSQLIdentifier_UnderscorePrefix` | `"_private"` | nil error | P2 |
| UT-008 | `TestToFloat64_Int` | `123` | `(123.0, nil)` | P0 |
| UT-009 | `TestToFloat64_String` | `"45.67"` | `(45.67, nil)` | P0 |
| UT-010 | `TestToFloat64_StringInvalid` | `"abc"` | `(0, error)` | P0 |
| UT-011 | `TestToFloat64_Bool` | `true` | `(0, error)` | P1 |
| UT-012 | `TestToFloat64_Int64` | `int64(99)` | `(99.0, nil)` | P1 |
| UT-013 | `TestToFloat64_Nil` | `nil` | `(0, error)` | P1 |
| UT-014 | `TestGormEqualFilter_ApplyGorm` | field="age", value=25 | SQL: `age = ?` | P0 |
| UT-015 | `TestGormFilter_SQLInjection` | field="id;DROP" | gorm.DB.Error != nil | P0 |
| UT-016 | `TestGormTranslatorRegistry_AllOperators` | 所有11种operator | 全部注册 | P1 |
| UT-017 | `TestGormTranslatorRegistry_UnsupportedOperator` | "unknown" | error | P1 |
| UT-018 | `TestRedisGreaterThanFilter_ApplyRedis` | field="age", value=30 | 过滤后 keys | P0 |
| UT-019 | `TestRedisGreaterThanFilter_Invalid` | value="abc" | error (toFloat64 fail) | P0 |
| UT-020 | `TestApplyRedisFilters_Multiple` | 3个filter链 | 正确过滤链式应用 | P1 |

## util/env

| ID | 测试函数 | 环境变量 | 预期输出 |
|----|---------|---------|---------|
| UT-101 | `TestGetCacheAsideTTL_CACHE_ASIDE_TTL` | `CACHE_ASIDE_TTL=60` | 60s |
| UT-102 | `TestGetCacheAsideTTL_CACHE_TTL_SECONDS` | `CACHE_TTL_SECONDS=120` | 120s |
| UT-103 | `TestGetCacheAsideTTL_Both` | 两变量都设 | CACHE_ASIDE_TTL 优先 |
| UT-104 | `TestGetCacheAsideTTL_Invalid` | `CACHE_ASIDE_TTL=abc` | 默认1h |
| UT-105 | `TestGetCacheAsideTTL_Negative` | `CACHE_ASIDE_TTL=-5` | 默认1h |
| UT-106 | `TestGetCacheAsideTTL_Zero` | `CACHE_ASIDE_TTL=0` | 默认1h |
| UT-107 | `TestGetCacheHitRefresh_True` | `CACHE_HIT_REFRESH=true` | true |
| UT-108 | `TestGetCacheHitRefresh_False` | 不设或设false | false |
| UT-109 | `TestGetEnvOrDefault_Found` | `PORT=9090` | "9090" |
| UT-110 | `TestGetEnvOrDefault_Default` | 无 | defaultValue |

## util/cache_key_builder

| ID | 测试函数 | 输入 | 预期输出 |
|----|---------|------|---------|
| UT-201 | `TestTemplateKeyBuilder_Normal` | template=`user:{id}`, data.ID=5 | "user:5" |
| UT-202 | `TestTemplateKeyBuilder_MultiField` | template=`product:{id}:{category}` | "product:1:electronics" |
| UT-203 | `TestTemplateKeyBuilder_Nested` | template=`cache:{user.id}` | 提取嵌套字段 |
| UT-204 | `TestTemplateKeyBuilder_NilData` | data=nil | 原始模板 |
| UT-205 | `TestTemplateKeyBuilder_JsonTag` | 按 json tag 匹配 | 正确匹配 |
| UT-206 | `TestPrefixKeyBuilder` | prefix="user", id=10 | "user:10" |

## service (纯逻辑方法)

| ID | 测试函数 | 描述 |
|----|---------|------|
| UT-301 | `TestGetTypeName_Struct` | struct{} → "struct{}" |
| UT-302 | `TestGetTypeName_Pointer` | *User → "User" |
| UT-303 | `TestNewServiceManager_Defaults` | TableName == ResourceName |
| UT-304 | `TestExtractID_Valid` | JSON 包含 id 字段 |
| UT-305 | `TestExtractID_NoIDField` | JSON 无 id 字段 → ok=false |
| UT-306 | `TestExtractIDFromKey_Valid` | "user:123" → (123, nil) |
| UT-307 | `TestExtractIDFromKey_NoColon` | "simple" → error |
| UT-308 | `TestExtractIDFromKey_NonNumeric` | "user:abc" → error |
| UT-309 | `TestExtractIDFromKey_TripleColon` | "cache:user:456" → (456, nil) |
| UT-310 | `TestHavingCondition_SQLInjection` | Operator="; DROP TABLE" → error |
| UT-311 | `TestApplyQueryOptions_Order_Invalid` | Order="DELETE" → AddError |
