# AbstractManager 测试方案

> 版本: 1.0 | 更新日期: 2026-06-06 | 覆盖代码行数: ~4000+

## 目录

1. [测试策略总览](#1-测试策略总览)
2. [测试环境与基础设施](#2-测试环境与基础设施)
3. [单元测试方案](#3-单元测试方案)
4. [集成测试方案](#4-集成测试方案)
5. [竞态测试方案 (Race Detection)](#5-竞态测试方案-race-detection)
6. [性能基准测试方案 (Benchmark)](#6-性能基准测试方案-benchmark)
7. [功能测试方案 (端到端)](#7-功能测试方案-端到端)
8. [测试覆盖率目标](#8-测试覆盖率目标)
9. [CI/CD 集成](#9-cicd-集成)

---

## 1. 测试策略总览

### 1.1 测试金字塔

```
         ╱  E2E  ╲          ~10 个场景，使用 docker-compose 真实环境
        ╱ 集成测试 ╲         ~40 个场景，miniredis + sqlite 内存环境
       ╱  竞态测试  ╲        race detector + `go test -race` 全覆盖
      ╱   Benchmark  ╲      关键热点 benchmark + pprof 分析
     ╱   单元测试     ╲      每个包独立可运行，mock 外部依赖
```

### 1.2 目标矩阵

| 层 | 覆盖率目标 | 运行时间 | 工具 |
|----|-----------|---------|------|
| 单元测试 | ≥ 80% | < 30s | `go test` |
| 集成测试 | ≥ 70% | < 120s | `go test`, miniredis, go-sqlite3 |
| 竞态测试 | 100% 覆盖 goroutine 路径 | < 60s | `go test -race` |
| 性能测试 | 关键路径 P99 < 10ms | < 300s | `go test -bench`, pprof |
| 功能测试 | 核心业务流程全覆盖 | < 300s | docker-compose, Go HTTP client |

---

## 2. 测试环境与基础设施

### 2.1 依赖矩阵

| 组件 | 单元测试 | 集成测试 | E2E 测试 |
|------|---------|---------|---------|
| MySQL | mock (go-sqlmock) | SQLite in-memory | Docker mysql:8.0 |
| Redis | mock (miniredis) | miniredis v2 | Docker redis:7-alpine |
| HTTP | httptest.ResponseRecorder | Gin test mode | Docker + real HTTP |
| 时间 | mock clock | time.Now() | 真实时间 |

### 2.2 推荐的测试依赖

```go
// go.mod 测试依赖
require (
    github.com/alicebob/miniredis/v2 v2.34.0   // Redis mock (纯 Go 实现)
    github.com/DATA-DOG/go-sqlmock v1.5.2        // SQL mock
    github.com/mattn/go-sqlite3 v1.14.22         // SQLite for integration
    github.com/stretchr/testify v1.10.0          // assertions + mock
    gorm.io/driver/sqlite v1.5.6                 // GORM SQLite driver
)
```

### 2.3 测试辅助目录结构

```
AbstractManager/
├── testutil/                          # 共享测试工具
│   ├── mock_redis.go                  # miniredis 启动/停止 helper
│   ├── mock_db.go                     # SQLite in-memory helper
│   ├── fixtures.go                    # 测试数据工厂
│   └── assertions.go                  # 自定义断言
├── service/
│   ├── *_test.go                      # 同包测试
│   └── testdata/                      # 测试数据 (json)
├── util/
│   ├── *_test.go
│   └── testdata/
├── http_router/
│   ├── *_test.go
│   └── testdata/
└── docs/
    └── test_plan/                     # 本测试方案
```

---

## 3. 单元测试方案

### 3.1 util/filter_translator 包

**文件**: `util/filter_translator/filter_test.go` (已有雏形，需扩展)

```
测试文件: util/filter_translator/filter_test.go
测试文件: util/filter_translator/gorm_filter_test.go
测试文件: util/filter_translator/redis_filter_test.go
```

#### 3.1.1 `ValidateSQLIdentifier` 测试

```go
// filter_validate_test.go
func TestValidateSQLIdentifier(t *testing.T)
├── TestValidateSQLIdentifier_Valid      // 合法标识符: "id", "user_name", "_private"
├── TestValidateSQLIdentifier_Empty      // 空字符串
├── TestValidateSQLIdentifier_Invalid    // SQL 注入: "1;DROP TABLE", "id--", "id/*"
├── TestValidateSQLIdentifier_NumericPrefix // "1invalid"
└── TestValidateSQLIdentifier_SpecialChars  // "id@name", "user space"
```

#### 3.1.2 `toFloat64` 测试

```go
func TestToFloat64(t *testing.T)
├── TestToFloat64_Int
├── TestToFloat64_Int64
├── TestToFloat64_Float64
├── TestToFloat64_StringValid
├── TestToFloat64_StringInvalid       // "not_a_number"
└── TestToFloat64_UnknownType         // struct{}, bool
```

#### 3.1.3 GORM Filter 单元测试 (`gorm_filter_test.go`)

```go
// 使用 sqlmock 验证 SQL 生成正确性
func TestGormEqualFilter_ApplyGorm(t *testing.T)
func TestGormInFilter_ApplyGorm(t *testing.T)
func TestGormBetweenFilter_ApplyGorm(t *testing.T)
func TestGormLikeFilter_ApplyGorm(t *testing.T)
func TestGormIsNullFilter_ApplyGorm(t *testing.T)
func TestGormFilter_SQLInjection(t *testing.T)  // 注入字段名应被拦截

// 翻译器单元测试
func TestGormEqualTranslator_Translate(t *testing.T)
func TestGormEqualTranslator_Validate(t *testing.T)
func TestGormInTranslator_Validate_EmptyArray(t *testing.T)
func TestGormBetweenTranslator_Validate_WrongArrayLen(t *testing.T)
func TestGormTranslatorRegistry_UnsupportedOperator(t *testing.T)
func TestGormTranslatorRegistry_TranslateBatch(t *testing.T)
```

#### 3.1.4 Redis Filter 单元测试 (`redis_filter_test.go`)

```go
// 使用 miniredis 模拟 Redis 行为
func TestRedisEqualFilter_ApplyRedis(t *testing.T)
func TestRedisGreaterThanFilter_ApplyRedis(t *testing.T)
func TestRedisGreaterThanFilter_InvalidValue(t *testing.T) // 期望返回 error
func TestRedisLessThanFilter_ApplyRedis(t *testing.T)
func TestRedisBetweenFilter_ApplyRedis(t *testing.T)
func TestRedisBetweenFilter_InvalidMinMax(t *testing.T)
func TestRedisInFilter_ApplyRedis(t *testing.T)
func TestRedisLikeFilter_ApplyRedis(t *testing.T)
func TestApplyRedisFilters_Multiple(t *testing.T)
```

### 3.2 util/env 包

**文件**: `util/env_test.go`

```go
func TestGetCacheAsideTTL(t *testing.T)
├── TestGetCacheAsideTTL_Default        // 未设环境变量，返回默认1h
├── TestGetCacheAsideTTL_CacheAsideTTL  // CACHE_ASIDE_TTL=60 → 60s
├── TestGetCacheAsideTTL_LegacyVar      // CACHE_TTL_SECONDS=120 → 120s
├── TestGetCacheAsideTTL_Priority       // 两个都设，CACHE_ASIDE_TTL 优先
└── TestGetCacheAsideTTL_InvalidValue   // "abc" → 默认

func TestGetCacheHitRefresh(t *testing.T)
├── TestGetCacheHitRefresh_True
└── TestGetCacheHitRefresh_False

func TestGetEnvOrDefault(t *testing.T)
├── TestGetEnvOrDefault_Exists
└── TestGetEnvOrDefault_Default
```

### 3.3 util/cache_key_builder 包

**文件**: `util/cache_key_builder/cache_key_builder_test.go`

```go
type TestUser struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Role     string `json:"role"`
}

func TestTemplateKeyBuilder_BuildKey(t *testing.T)
├── TestTemplateKeyBuilder_Simple
├── TestTemplateKeyBuilder_NestedField  // {user.id}
├── TestTemplateKeyBuilder_CaseInsensitive
├── TestTemplateKeyBuilder_JsonTag
└── TestTemplateKeyBuilder_NilData

func TestFuncKeyBuilder_BuildKey(t *testing.T)
func TestPrefixKeyBuilder_BuildKey(t *testing.T)
func TestQuickBuildKey(t *testing.T)
```

### 3.4 service 包 — 纯逻辑方法的单元测试

以下方法可以用 sqlmock 或纯 mock 独立测试：

```go
// service_model_test.go
func TestGetTypeName(t *testing.T)
├── TestGetTypeName_Struct
├── TestGetTypeName_Pointer
└── TestGetTypeName_Primitive

func TestNewServiceManager(t *testing.T)
├── TestNewServiceManager_Defaults         // 验证自动推导的表名/键名
└── TestNewServiceManager_CustomSchema
```

```go
// get_query_test.go (纯 SQL 生成逻辑，用 sqlmock)
func TestApplyQueryOptions(t *testing.T)
├── TestApplyQueryOptions_Nil
├── TestApplyQueryOptions_Pagination
├── TestApplyQueryOptions_OrderBy_Asc
├── TestApplyQueryOptions_OrderBy_Desc
├── TestApplyQueryOptions_OrderBy_InvalidDirection
├── TestApplyQueryOptions_Group
├── TestApplyQueryOptions_Having_Legacy
├── TestApplyQueryOptions_HavingConditions
├── TestApplyQueryOptions_Having_InvalidOperator
├── TestApplyQueryOptions_Distinct
├── TestApplyQueryOptions_Select
└── TestApplyQueryOptions_Preload

func TestHavingCondition_OperatorWhitelist(t *testing.T)
```

```go
// lookup_query_test.go (纯逻辑方法单元测试)
func TestExtractID(t *testing.T)
├── TestExtractID_Valid
├── TestExtractID_NoIDField
└── TestExtractID_NilInput

func TestExtractIDFromKey(t *testing.T)
├── TestExtractIDFromKey_Valid          // "user:123" → 123
├── TestExtractIDFromKey_NoColon        // "simplekey"
├── TestExtractIDFromKey_NonNumeric     // "user:abc"
└── TestExtractIDFromKey_MultipleColons // "cache:user:456" → 456
```

```go
// writedown_query_test.go (逻辑测试)
func TestWritedownQuery_EmptyData(t *testing.T)
func TestWritedownQuery_NilOpts(t *testing.T)       // 验证默认值
func TestWritedownQuery_CustomBatchSize(t *testing.T)
```

### 3.5 http_router 包 — Handler 单元测试

```go
// get_router_group_test.go
func TestQueryRouterGroup_HandleQuery_InvalidJSON(t *testing.T)
func TestQueryRouterGroup_HandleQuery_UnknownMethod(t *testing.T)
func TestQueryRouterGroup_HandleCount(t *testing.T)
func TestQueryRouterGroup_HandleGetByID_NotFound(t *testing.T)
```

```go
// cache_get_router_group_test.go
func TestLookupRouterGroup_HandleLookup_InvalidJSON(t *testing.T)
func TestLookupRouterGroup_HandleLookup_MissingKeyPattern(t *testing.T)
func TestLookupRouterGroup_HandleInvalidate_NoKeysOrPattern(t *testing.T)
func TestDeriveKeyPrefix(t *testing.T)
├── TestDeriveKeyPrefix_Struct          // model.User → "user"
└── TestDeriveKeyPrefix_Pointer         // *model.User → "user"
```

---

## 4. 集成测试方案

集成测试使用 **miniredis** (内存Redis) + **SQLite** (内存数据库) 替代真实服务，确保测试可并行、无外部依赖。

### 4.1 测试基础架构

```go
// testutil/integration_helper.go
package testutil

import (
    "github.com/alicebob/miniredis/v2"
    "github.com/redis/go-redis/v9"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// IntegrationEnv 集成测试环境
type IntegrationEnv struct {
    DB          *gorm.DB
    Redis       *redis.Client
    MiniRedis   *miniredis.Miniredis  // 用于直接控制时间/数据
    Service     *service.ServiceManager[TestUser]
    Cleanup     func()
}

// SetupIntegrationEnv 创建集成测试环境
func SetupIntegrationEnv(t *testing.T) *IntegrationEnv {
    // 1. SQLite 内存模式
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    service.SetTestDB(db)  // 注入测试 DB

    // 2. miniredis
    mr := miniredis.RunT(t)
    rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
    service.SetTestRedis(rdb)

    // 3. 创建 ServiceManager
    sm := service.NewServiceManager(TestUser{})

    return &IntegrationEnv{
        DB:        db,
        Redis:     rdb,
        MiniRedis: mr,
        Service:   sm,
    }
}

type TestUser struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Username string `gorm:"uniqueIndex" json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
    Status   string `json:"status"`
}
```

### 4.2 service 层集成测试

#### 4.2.1 DDL 操作 (`create_integration_test.go`)

```go
func TestCreate_Integration(t *testing.T)
├── TestCreate_IfNotExists_FirstTime    // 首次创建成功
├── TestCreate_IfNotExists_Repeated     // 重复创建不报错
├── TestCreate_DropIfExists             // 删除后重建
├── TestCreateWithIndexes               // 带索引创建
├── TestCreateIndex_Unique              // 唯一索引
├── TestCreateIndex_NonUnique           // 普通索引
├── TestDropTable                       // 删除表
└── TestHasTable                        // 检查表存在
```

#### 4.2.2 写入操作 (`set_integration_test.go`)

```go
func TestSetSingle_Integration(t *testing.T)
├── TestSetSingle_Insert
├── TestSetSingle_Upsert                // OnConflict 更新
├── TestSetSingle_Insert_NoConflict     // 冲突时跳过
├── TestInsert                          // 封装方法
├── TestUpdate                          // 条件更新
├── TestUpdateByID                      // 按 ID 更新
├── TestSave                            // GORM Save
├── TestUpsert                          // 指定冲突列
├── TestDelete                          // 条件删除
├── TestDeleteByID                      // 按 ID 删除
├── TestSoftDelete                      // 软删除
├── TestSoftDeleteByID                  // 按 ID 软删除
├── TestIncrement                       // 字段自增
├── TestIncrementByID                   // 按 ID 自增
├── TestDecrement                       // 字段自减
└── TestDecrementByID                   // 按 ID 自减
```

#### 4.2.3 查询操作 (`get_integration_test.go`)

```go
func TestGetSingle_Integration(t *testing.T)
├── TestGetSingle_Found
├── TestGetSingle_NotFound
├── TestGetSingleByID
├── TestGetSingleOrCreate_Exists        // 存在
├── TestGetSingleOrCreate_NotExists     // 不存在，自动创建
├── TestGetSingleWithLock
├── TestGetFirst
├── TestGetLast
└── TestGetSingle_Preload
```

```go
func TestGetQuery_Integration(t *testing.T)
├── TestGetQuery_NoFilter
├── TestGetQuery_Pagination
├── TestGetQuery_PageZero               // page=0 → 默认 page=1
├── TestGetQuery_OrderBy
├── TestGetQuery_OrderBy_Desc
├── TestGetQuery_Group
├── TestGetQuery_HavingLegacy
├── TestGetQuery_HavingConditions       // 结构化 Having
├── TestGetQuery_Having_GreaterThan
├── TestGetQuery_Distinct
├── TestGetQuery_Select
├── TestGetQuery_Preload
├── TestCountQuery
├── TestExistsQuery_True
└── TestExistsQuery_False
```

#### 4.2.4 批量操作 (`batch_integration_test.go`)

```go
func TestSetQuery_Integration(t *testing.T)
├── TestSetQuery_Insert
├── TestSetQuery_Upsert
├── TestSetQuery_Empty
├── TestSetQuery_BatchSplit            // 大于 BatchSize 时自动分批
├── TestBatchInsert                    // 非冲突插入
├── TestBatchUpdate                    // 批量条件更新
├── TestBatchUpsert                    // 指定冲突列
├── TestBatchDelete                    // 批量条件删除
├── TestBatchSoftDelete                // 批量软删除
├── TestBatchIncrement                 // 批量自增
├── TestBatchDecrement                 // 批量自减
└── TestSetQuery_InvalidateCache       // 写入后缓存失效
```

#### 4.2.5 缓存读写 (`cache_integration_test.go`)

```go
func TestWritedownSingle_Integration(t *testing.T)
├── TestWritedownSingle_Set
├── TestWritedownSingle_SetNX          // 仅不存在时设置
├── TestWritedownSingle_SetXX          // 仅存在时设置
├── TestWritedownSingle_Overwrite      // 覆盖
├── TestWritedownSingleByID            // 从 DB 读取后写入缓存
├── TestWritedownSingleWithVersion     // 版本控制
├── TestWritedownSingleWithVersion_Outdated  // 版本过旧被拒绝
└── TestWritedownSingleWithLock        // 分布式锁写缓存

func TestWritedownQuery_Integration(t *testing.T)
├── TestWritedownQuery_Batch
├── TestWritedownQuery_OverwriteFalse
├── TestWritedownQuery_MarshalError
├── TestWritedownQuery_BatchSplit      // 验证 Pipeline 分片
├── TestWritedownQueryFromDB
├── TestWritedownQueryByIDs
├── TestWritedownAllToCache
└── TestWarmupCache

func TestLookupSingle_Integration(t *testing.T)
├── TestLookupSingle_CacheHit
├── TestLookupSingle_CacheMiss
├── TestLookupSingleWithFallback       // 缓存未命中回源 DB
├── TestLookupSingleByID               // 按 ID 查询
├── TestInvalidateSingleCache          // 失效单个
├── TestInvalidateSingleCacheByID
├── TestExistsInCache_True
├── TestExistsInCache_False
├── TestExtendCacheTTL
└── TestGetCacheTTL
```

```go
func TestLookupQuery_Integration(t *testing.T)
├── TestLookupQuery_AllHit
├── TestLookupQuery_PartialHit          // 部分命中，回源
├── TestLookupQuery_AllMiss_Fallback
├── TestLookupQuery_AllMiss_NoFallback
├── TestLookupQuery_EmptyKeys
├── TestLookupQuery_InvalidJSON         // 缓存中有损坏的 JSON
├── TestLookupQuery_KeyExtraction       // 从 key 提取 ID 正确
├── TestLookupQueryByPattern
├── TestLookupQueryWithRefresh
├── TestRefreshCache
├── TestInvalidateCache
└── TestInvalidateCacheByPattern
```

#### 4.2.6 异步写入 (`async_integration_test.go`)

```go
func TestWritedownSingleAsync_Integration(t *testing.T)
├── TestAsync_Success                   // 异步写入成功
├── TestAsync_QueueFull                  // 队列满时丢弃
├── TestAsync_WorkerStartup              // worker 惰性启动
├── TestAsync_WorkersStartOnlyOnce       // 重复调用只启动一次
├── TestAsync_ShutdownGraceful           // 优雅关闭，排空队列
└── TestAsync_ShutdownMultiple           // 重复关闭不 panic
```

### 4.3 http_router 集成测试

使用 Gin 的 `httptest` + 注入 miniredis/SQLite 的 ServiceManager：

```go
// cache_get_router_group_integration_test.go
func TestLookupRouterGroup_Integration(t *testing.T)
├── TestHandleLookup_EmptyCache         // Redis 空，从 DB 加载
├── TestHandleLookup_WithFilters        // 带过滤条件
├── TestHandleLookup_FallbackDB         // 缓存未命中回源
├── TestHandleGetByKey_CacheHit         // 缓存命中
├── TestHandleGetByKey_CacheMiss_Fallback // 未命中回源
├── TestHandleGetByKey_NotFound         // 不存在
├── TestHandleGetByKey_RefreshTTL       // 命中后刷新 TTL
├── TestHandleCount                     // 计数
├── TestHandleInvalidate_ByKeys         // 精确删除
└── TestHandleInvalidate_ByPattern      // 模式删除
```

```go
// get_router_group_integration_test.go
func TestQueryRouterGroup_Integration(t *testing.T)
├── TestHandleQuery_List
├── TestHandleQuery_Search
├── TestHandleQuery_ActiveList
├── TestHandleGetByID_Found
├── TestHandleGetByID_NotFound
└── TestHandleCount
```

### 4.4 RedisManager & DBManager 集成测试

```go
// cache_pool_integration_test.go
func TestRedisManager_Integration(t *testing.T)
├── TestSet_Struct
├── TestSet_Bytes
├── TestGet_Found
├── TestGet_NotFound
├── TestDelete
├── TestExists
├── TestSetMultiple                      // Pipeline 批量写入
├── TestGetMultiple                      // Pipeline 批量读取
├── TestScanKeys                         // SCAN 安全遍历
└── TestScanKeys_LargeDataset            // 大量 key 分页扫描
```

---

## 5. 竞态测试方案 (Race Detection)

Go 的 race detector 可以检测数据竞争。所有包含 goroutine 的代码必须在 `-race` 下通过。

### 5.1 异步 Worker Pool 竞态测试

```go
// service/writedown_single_race_test.go
// 运行: go test -race -run TestAsync -count=10 ./service/

func TestAsyncWorkers_Race_ConcurrentSubmit(t *testing.T)
// 场景: 100 个 goroutine 并发调用 WritedownSingleAsync，每个 10 次提交
// 检测: asyncStarted 读写竞争、channel 并发发送、WaitGroup 计数

func TestAsyncWorkers_Race_StartShutdown(t *testing.T)
// 场景: 同时调用 startAsyncWorkersOnce 和 ShutdownAsyncWorkers
// 检测: asyncMu 锁保护、asyncShutdown channel close 竞争

func TestAsyncWorkers_Race_HotLoop(t *testing.T)
// 场景: 4 个 goroutine 持续写入 + 2 个 goroutine 定期关闭/重启
// 持续 5 秒
// 检测: 长期运行的热竞争路径

func TestAsyncWorkers_Race_ShutdownWhileWriting(t *testing.T)
// 场景: worker 正在处理任务时收到 shutdown 信号
// 检测: 从 asyncTasks channel 读取 vs 排空队列的 select 分支
```

### 5.2 ServiceManager 全局状态竞态测试

```go
// service/service_model_race_test.go
func TestServiceManager_Race_ConcurrentAccess(t *testing.T)
// 场景: 多个 goroutine 同时读写 ServiceManager 字段
// 检测: Resource, TableName, CacheKeyName 等字段并发访问

func TestServiceManager_Race_MultipleInstances(t *testing.T)
// 场景: 创建多个 ServiceManager 实例并发操作不同表
// 检测: globalRedisManager/globalDBManager 读写的安全性
```

### 5.3 缓存操作竞态测试

```go
// service/cache_race_test.go
func TestCache_Race_WriteReadSameKey(t *testing.T)
// 场景: 10 个 writer + 20 个 reader 并发操作同一 key
// 检测: Redis Pipeline 并发安全、JSON marshal 无竞争

func TestCache_Race_InvalidateWhileReading(t *testing.T)
// 场景: LookupQuery 执行过程中，另一个 goroutine 调用 InvalidateCache
// 检测: 返回结果的一致性

func TestCache_Race_WritedownWithVersion(t *testing.T)
// 场景: 多个 goroutine 并发调用 WritedownSingleWithVersion
// 检测: Watch + TxPipelined 原子性保证
```

### 5.4 Filter Translator 竞态测试

```go
// util/filter_translator/filter_race_test.go
func TestTranslatorRegistry_Race_ConcurrentTranslate(t *testing.T)
// 场景: 20 个 goroutine 同时调用 Translate/TranslateBatch
// 检测: registry.translators map 并发读取安全

func TestDefaultRegistry_Race(t *testing.T)
// 场景: 多 goroutine 并发使用 DefaultGormRegistry/DefaultRedisRegistry
// 检测: 全局单例的读安全（不可变注册表应该安全）
```

### 5.5 竞态测试运行命令

```bash
# 全量竞态测试 (每次运行多次以增加检测概率)
go test -race -count=5 ./...

# 指定高风险的包 (运行更多次)
go test -race -count=20 ./service/...

# 长时间压力竞态测试
go test -race -count=1 -timeout=300s -run TestAsyncWorkers_Race_HotLoop ./service/

# 验证数据竞争为 0
go test -race -json ./... | grep '"race"' | wc -l   # 期望输出: 0
```

---

## 6. 性能基准测试方案 (Benchmark)

### 6.1 基准测试目录结构

```
service/
├── writedown_query_bench_test.go       # 批量写入 Benchmark
├── lookup_query_bench_test.go          # 批量查询 Benchmark
├── set_query_bench_test.go             # 批量写入 DB Benchmark
├── get_query_bench_test.go             # 批量查询 DB Benchmark
├── writedown_single_bench_test.go      # 单条写入对比 Benchmark
└── cache_pool_bench_test.go            # Redis 操作 Benchmark

util/filter_translator/
├── filter_bench_test.go                # 过滤性能 Benchmark
└── cache_key_builder/
    └── cache_key_builder_bench_test.go # 键构建 Benchmark
```

### 6.2 WritedownQuery Pipeline 性能

```go
// writedown_query_bench_test.go
func BenchmarkWritedownQuery_Pipeline_100(b *testing.B)
// 对比 Pipeline + Set(TTL) vs MSet + 逐个 Expire
// 100 条记录，期望 Pipeline 快 5-10x

func BenchmarkWritedownQuery_Pipeline_1000(b *testing.B)
// 1000 条记录

func BenchmarkWritedownQuery_Pipeline_10000(b *testing.B)
// 10000 条记录，验证大 batch 性能

func BenchmarkWritedownQuery_BatchSize_50(b *testing.B)
func BenchmarkWritedownQuery_BatchSize_100(b *testing.B)
func BenchmarkWritedownQuery_BatchSize_500(b *testing.B)
func BenchmarkWritedownQuery_BatchSize_1000(b *testing.B)
// 不同 BatchSize 对比，找到最佳批次大小
```

### 6.3 序列化性能

```go
// writedown_single_bench_test.go
func BenchmarkMarshalForRedis_JSON(b *testing.B)
// 标准 JSON 序列化

func BenchmarkMarshalForRedis_JSON_WithPool(b *testing.B)
// 对比是否使用 sync.Pool 缓存 encoder

func BenchmarkWritedownSingle_JSON(b *testing.B)
// 单条写入端到端延迟
```

### 6.4 过滤性能

```go
// filter_bench_test.go
func BenchmarkApplyRedisFilters_1Filter_100Keys(b *testing.B)
func BenchmarkApplyRedisFilters_5Filters_100Keys(b *testing.B)
func BenchmarkApplyRedisFilters_10Filters_1000Keys(b *testing.B)
// 过滤器链式应用性能

func BenchmarkExtractID_JSON(b *testing.B)
// extractID 通过 JSON 往返的性能
// 备注: 如果成为瓶颈，考虑代码生成替代
```

### 6.5 DB 操作性能

```go
// set_query_bench_test.go
func BenchmarkSetQuery_Batch100(b *testing.B)
func BenchmarkSetQuery_Batch1000(b *testing.B)
func BenchmarkBatchUpsert_100(b *testing.B)
func BenchmarkBatchInsert_100(b *testing.B)
```

### 6.6 Cache Aside 路径性能

```go
// lookup_single_bench_test.go
func BenchmarkLookupSingleWithFallback_CacheHit(b *testing.B)
// 纯缓存命中路径，期望 < 1ms

func BenchmarkLookupSingleWithFallback_CacheMiss(b *testing.B)
// 缓存未命中 + DB 查询 + 异步回填

func BenchmarkBuildCacheKey(b *testing.B)
// 键构建性能
```

### 6.7 并发压力基准

```go
// service/throughput_bench_test.go
func BenchmarkThroughput_ReadHeavy(b *testing.B)
// 80% 读 + 20% 写，模拟高并发读场景
// RunParallel 模式

func BenchmarkThroughput_WriteHeavy(b *testing.B)
// 20% 读 + 80% 写

func BenchmarkThroughput_Mixed(b *testing.B)
// 50% 读 + 50% 写
```

### 6.8 运行 Benchmark

```bash
# 单次基准测试
go test -bench=. -benchmem -benchtime=3s ./service/...
go test -bench=. -benchmem -benchtime=3s ./util/...

# 对比两种实现 (旧实现需临时还原)
go test -bench=BenchmarkWritedownQuery_Pipeline -benchmem -count=5 ./service/ > new.txt
go test -bench=BenchmarkWritedownQuery_MSetExpire -benchmem -count=5 ./service/ > old.txt
benchstat old.txt new.txt

# CPU profile
go test -bench=BenchmarkWritedownQuery_Pipeline_1000 -cpuprofile=cpu.prof ./service/
go tool pprof -http=:8080 cpu.prof

# Memory profile
go test -bench=BenchmarkMarshalForRedis -memprofile=mem.prof ./service/
go tool pprof -http=:8081 mem.prof
```

---

## 7. 功能测试方案 (端到端)

### 7.1 docker-compose 环境

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: test
      MYSQL_DATABASE: abstract_manager_test
    ports: ["3307:3306"]
  redis:
    image: redis:7-alpine
    ports: ["6380:6379"]
```

### 7.2 E2E 测试用例

```go
// e2e/cache_aside_e2e_test.go
// 运行: go test -tags=e2e ./e2e/

func TestCacheAside_ReadThrough(t *testing.T)
// 1. 请求数据（首次 → 缓存未命中 → 回源 DB → 写入缓存）
// 2. 再次请求（缓存命中 → 返回缓存）
// 3. 验证 TTL 设置正确
// 4. 验证缓存命中返回 source=cache

func TestCacheAside_WriteThrough(t *testing.T)
// 1. 写入数据到 DB
// 2. 验证对应缓存已失效
// 3. 下次读取触发 Read Through

func TestCacheAside_BulkSync_NoRecache(t *testing.T)
// 1. 在 Redis 中创建 100 条缓存
// 2. 调用同步 API 写入 DB
// 3. 验证 DB 中有 100 条数据
// 4. 验证 RecacheAfterSync=false 时缓存不残留

func TestCacheAside_PeriodicSync(t *testing.T)
// 1. 启动定时同步（间隔 2s）
// 2. 创建缓存数据
// 3. 等待 5s
// 4. 验证 DB 中数据已同步

func TestCacheAside_ConcurrentReadWrite(t *testing.T)
// 1. 后台 goroutine 持续写入数据（每 100ms 一条）
// 2. 后台 goroutine 持续读取（每 50ms 一次）
// 3. 运行 30s
// 4. 验证无数据丢失、无 panic
```

```go
// e2e/api_e2e_test.go
func TestAPI_LookupEndpoint(t *testing.T)
// POST /api/v1/users/lookup/lookup

func TestAPI_GetByKeyEndpoint(t *testing.T)
// GET /api/v1/users/lookup/:key

func TestAPI_CountEndpoint(t *testing.T)
// POST /api/v1/users/lookup/count

func TestAPI_InvalidateEndpoint(t *testing.T)
// POST /api/v1/users/lookup/invalidate

func TestAPI_WritedownEndpoint(t *testing.T)
// POST /api/v1/users/cache/*

func TestAPI_QueryEndpoint(t *testing.T)
// POST /api/v1/users/query

func TestAPI_GetByIDEndpoint(t *testing.T)
// GET /api/v1/users/:id

func TestAPI_GracefulShutdown(t *testing.T)
// 1. 启动 server
// 2. 发送请求
// 3. 发送 SIGTERM
// 4. 验证正在进行的请求能完成
// 5. 验证异步 worker 排空队列后退出
```

### 7.3 边界场景与异常测试

```go
// e2e/failure_scenarios_e2e_test.go
func TestRedisConnectionFailure(t *testing.T)
// Redis 断开时，DB 查询正常返回（不崩溃）

func TestDBConnectionFailure(t *testing.T)
// DB 断开时，Redis 缓存命中仍正常返回

func TestLargePayload(t *testing.T)
// 单条数据 > 1MB 时序列化和缓存写入

func TestEmptyDataScenarios(t *testing.T)
// 空数组、空 map、零值结构体

func TestUnicodeData(t *testing.T)
// 包含中文、emoji 的数据正确序列化往返

func TestConcurrentTableCreate(t *testing.T)
// 多 goroutine 同时 Create 同一张表

func TestCacheExpiration(t *testing.T)
// TTL 过期后缓存自动清除
// 使用 miniredis.FastForward() 模拟时间流逝

func TestForUpdate_LockTimeout(t *testing.T)
// GetSingle 的 ForUpdate 锁竞争
```

---

## 8. 测试覆盖率目标

### 8.1 分模块覆盖率

| 模块 | 目标 | 关键覆盖点 |
|------|------|-----------|
| `util/filter_translator` | ≥ 90% | 所有 Filter 类型、所有 Translator、错误路径 |
| `util/env` | ≥ 90% | 环境变量存在/不存在/非法值 |
| `util/cache_key_builder` | ≥ 85% | 三种 KeyBuilder 实现 |
| `service/create.go` | ≥ 80% | 建表、建索引、删表 |
| `service/get_query.go` | ≥ 80% | QueryOptions 所有分支 |
| `service/get_single.go` | ≥ 80% | ForUpdate、GetSingleOrCreate |
| `service/set_query.go` | ≥ 80% | 批量 Upsert、BatchIncrement |
| `service/set_single.go` | ≥ 80% | Upsert、SoftDelete |
| `service/writedown_single.go` | ≥ 85% | 异步 worker、版本控制、分布式锁 |
| `service/writedown_query.go` | ≥ 80% | Pipeline 批量写入 |
| `service/lookup_query.go` | ≥ 80% | 批量查询、回源、key 解析 |
| `service/lookup_single.go` | ≥ 75% | 单个查询、TTL 操作 |
| `service/cache_pool.go` | ≥ 75% | Pipeline、ScanKeys |
| `service/sql_pool.go` | ≥ 60% | 初始化逻辑（需真实 DB 部分少） |
| `http_router/*` | ≥ 70% | Handler 请求处理、错误响应 |

### 8.2 运行覆盖率

```bash
# 全量覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 分模块覆盖率
go test -cover ./service/
go test -cover ./util/...
go test -cover ./http_router/...

# 排除 e2e 测试（它们使用 docker）
go test -coverprofile=coverage.out $(go list ./... | grep -v /e2e/)
```

---

## 9. CI/CD 集成

### 9.1 GitHub Actions 配置

```yaml
# .github/workflows/test.yml
name: Test
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  unit-and-integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: go test -coverprofile=unit.out -short ./...
      - run: go tool cover -func=unit.out

  race-detection:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: go test -race -count=3 -short ./...

  benchmark-compare:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: |
          git stash
          go test -bench=. -benchmem -count=5 ./... > base.txt
          git stash pop
          go test -bench=. -benchmem -count=5 ./... > head.txt
      - uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: go
          output-file-path: head.txt
          external-data-json-path: ./benchmarks.json

  e2e:
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env: { MYSQL_ROOT_PASSWORD: test, MYSQL_DATABASE: test }
        ports: ['3306:3306']
      redis:
        image: redis:7-alpine
        ports: ['6379:6379']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -tags=e2e -v ./e2e/
```

### 9.2 Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
echo "Running tests..."
go test -short ./...
if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running race detector..."
go test -race -short ./...
if [ $? -ne 0 ]; then
    echo "Race detector found issues. Commit aborted."
    exit 1
fi
```

---

## 附录 A: 测试辅助函数模板

```go
// testutil/fixtures.go
package testutil

type TestUser struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Username string `gorm:"uniqueIndex" json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
    Status   string `json:"status"`
}

func NewTestUser(id uint, username string) TestUser {
    return TestUser{
        ID:       id,
        Username: username,
        Email:    username + "@test.com",
        Age:      25,
        Status:   "active",
    }
}

func GenerateTestUsers(count int) []TestUser {
    users := make([]TestUser, count)
    for i := 0; i < count; i++ {
        users[i] = NewTestUser(uint(i+1), fmt.Sprintf("user_%d", i+1))
    }
    return users
}

func BuildKeyFunc() func(*TestUser) string {
    return func(u *TestUser) string {
        return fmt.Sprintf("user:%d", u.ID)
    }
}
```

## 附录 B: 快速开始

```bash
# 1. 安装测试依赖
go get -t ./...

# 2. 运行单元测试 (无外部依赖)
go test -short ./...

# 3. 运行集成测试 (miniredis + SQLite 内存模式)
go test -run Integration ./...

# 4. 运行竞态检测
go test -race -count=3 ./...

# 5. 运行性能基准
go test -bench=. -benchmem ./...

# 6. 运行 E2E (需要 Docker)
docker-compose -f docker-compose.test.yml up -d
go test -tags=e2e -v ./e2e/
docker-compose -f docker-compose.test.yml down

# 7. 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
