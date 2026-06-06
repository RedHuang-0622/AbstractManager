# 集成测试用例详细清单

> 使用 miniredis (内存 Redis) + SQLite in-memory 内存数据库，零外部依赖。

## 测试模型

```go
type TestUser struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Username string `gorm:"uniqueIndex;size:64" json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
    Status   string `gorm:"default:active" json:"status"`
}
```

## 集成测试用例

### DDL 操作

| ID | 测试 | 步骤 | 预期 |
|----|------|------|------|
| IT-001 | Create_IfNotExists_First | 1. sm.Create(ctx, {IfNotExists:true}) | err == nil, 表存在 |
| IT-002 | Create_IfNotExists_Repeat | 1. Create 2. Create 同一张表 | 第二次不报错 |
| IT-003 | CreateWithIndexes_Unique | 1. CreateWithIndexes({Unique:true}) 2. 插入两条相同索引值 | 第二次失败 |
| IT-004 | CreateWithIndexes_NonUnique | 1. 创建非唯一索引 2. 插入两条相同值 | 都成功 |

### 单条写入

| ID | 测试 | 描述 |
|----|------|------|
| IT-101 | SetSingle_Insert | 插入新记录 |
| IT-102 | SetSingle_Upsert | 插入 → 再插入相同 ID → 字段被更新 |
| IT-103 | SetSingle_Insert_NoConflict | 插入 → 再插入相同 ID → 第二次失败 |
| IT-104 | Update | 条件更新 age = 30 |
| IT-105 | UpdateByID | 按 ID 更新 username |
| IT-106 | Save | GORM Save (全字段覆盖) |
| IT-107 | Upsert | 指定冲突列的自定义 Upsert |
| IT-108 | Delete | 条件删除 |
| IT-109 | DeleteByID | 按 ID 删除 |
| IT-110 | SoftDelete | 软删除 → deleted_at 不为 null |
| IT-111 | Increment | age = age + 1 |
| IT-112 | IncrementByID | 按 ID 自增 |
| IT-113 | Decrement | age = age - 1 |
| IT-114 | DecrementByID | 按 ID 自减 |

### 单条查询

| ID | 测试 | 描述 |
|----|------|------|
| IT-201 | GetSingle_Found | 查询存在的记录 |
| IT-202 | GetSingle_NotFound | 返回 "record not found" |
| IT-203 | GetSingleByID | 按主键查询 |
| IT-204 | GetSingleOrCreate_Exists | 存在 → 返回已有 |
| IT-205 | GetSingleOrCreate_NotExists | 不存在 → 自动创建 |
| IT-206 | GetSingleWithLock | 加锁查询 → 另一事务无法更新 |
| IT-207 | GetFirst | 按 created_at ASC 首条 |
| IT-208 | GetLast | 按 created_at DESC 末条 |
| IT-209 | GetSingle_Preload | 带预加载关联 |

### 批量操作

| ID | 测试 | 描述 |
|----|------|------|
| IT-301 | SetQuery_Insert_50 | 批量插入 50 条 |
| IT-302 | SetQuery_Upsert_50 | 批量 Upsert 50 条 |
| IT-303 | SetQuery_BatchSplit_250 | 250条 + BatchSize=100 → 自动分为3批 |
| IT-304 | SetQuery_Empty | 空数组 → 直接返回 nil |
| IT-305 | BatchUpdate | 批量条件更新 status=inactive |
| IT-306 | BatchUpsert | 指定冲突列的批量 Upsert |
| IT-307 | BatchDelete | 批量条件删除 |
| IT-308 | BatchIncrement | 批量 age++ |
| IT-309 | BatchDecrement | 批量 age-- |

### 缓存写入

| ID | 测试 | 描述 |
|----|------|------|
| IT-401 | WritedownSingle_Set | 写入 → 读取验证 |
| IT-402 | WritedownSingle_SetNX | SetNX → 首次成功，第二次失败 |
| IT-403 | WritedownSingle_SetXX | SetXX → 不存在时失败 |
| IT-404 | WritedownSingle_Overwrite | 写入A → 覆盖写入B → 读B |
| IT-405 | WritedownSingleByID | 从DB查 → 写入缓存 |
| IT-406 | WritedownSingleWithVersion | V1写入 → V2写入成功 → V1被拒绝 |
| IT-407 | WritedownSingleWithLock | 锁竞争场景 |
| IT-408 | WritedownQuery_Batch_100 | 100条 Pipeline 写入 → 验证全部存在 |
| IT-409 | WritedownQuery_OverwriteFalse | 已存在 → 不覆盖 |
| IT-410 | WritedownQuery_BatchSplit | 250条+BatchSize=100 → 3个Pipeline |
| IT-411 | WritedownQueryFromDB | 从DB查询 → 全量写入缓存 |
| IT-412 | WritedownQueryByIDs | 按ID列表从DB查询 → 写入缓存 |
| IT-413 | WarmupCache | 预热缓存（从DB加载最新1000条） |

### 缓存查询

| ID | 测试 | 描述 |
|----|------|------|
| IT-501 | LookupSingle_CacheHit | 缓存存在 → 返回缓存 |
| IT-502 | LookupSingle_CacheMiss | 缓存不存在 → redis.Nil |
| IT-503 | LookupSingleWithFallback | 缓存miss → DB查询 → 异步回填 |
| IT-504 | LookupSingleByID | 自动构建 key 查询 |
| IT-505 | InvalidateSingleCache | 删除缓存 → 验证不存在 |
| IT-506 | ExistsInCache_True | 存在 → true |
| IT-507 | ExistsInCache_False | 不存在 → false |
| IT-508 | ExtendCacheTTL | 设置TTL → 延长TTL → 验证 |
| IT-509 | GetCacheTTL | 获取剩余TTL |
| IT-510 | LookupQuery_AllHit | 5个key全命中 |
| IT-511 | LookupQuery_PartialHit | 5个中3个命中 → 回源2个 |
| IT-512 | LookupQuery_AllMiss_Fallback | 全miss → 从DB加载 |
| IT-513 | LookupQuery_AllMiss_NoFallback | 全miss → 返回空 |
| IT-514 | LookupQuery_CorruptJSON | 缓存中有非法JSON → 跳过或报错 |
| IT-515 | LookupQueryByPattern | 按模式扫描 key 后查询 |
| IT-516 | LookupQueryWithRefresh | miss → DB+缓存 → 再查命中 |
| IT-517 | InvalidateCache | 删除指定keys → 验证全部不存在 |
| IT-518 | InvalidateCacheByPattern | 按模式删除 → 验证全部清理 |

### 异步 Worker

| ID | 测试 | 描述 |
|----|------|------|
| IT-601 | Async_Success | 提交任务 → 等待 → 缓存写入 |
| IT-602 | Async_WorkerStartOnce | 多次调用 → worker 只启动一次 |
| IT-603 | Async_QueueFull | 提交 300 个任务 → 部分被丢弃 |
| IT-604 | Async_ShutdownGraceful | 提交 50 个 → 立即 shutdown → 全部完成 |
| IT-605 | Async_ShutdownIdempotent | 调用两次 ShutdownAsyncWorkers → 不 panic |

### RedisManager & ScanKeys

| ID | 测试 | 描述 |
|----|------|------|
| IT-701 | Set_Struct | 写入结构体 → 读取 |
| IT-702 | Set_Bytes | 写入 []byte → 读取 |
| IT-703 | Get_NotFound | redis.Nil |
| IT-704 | SetMultiple_100 | 100个key批量Pipeline写入 |
| IT-705 | GetMultiple_50 | 50个key批量Pipeline读取 |
| IT-706 | ScanKeys_Small | 10个key → SCAN返回全部 |
| IT-707 | ScanKeys_Large | 200个key → SCAN多轮分页 |
| IT-708 | ScanKeys_Empty | 无匹配key → 返回空 |
