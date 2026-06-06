# 竞态与性能测试用例

> **实现文件:** [tests/race_and_perf_test.go](../../tests/race_and_perf_test.go)
>
> **实现状态:** ✅ 全部完成

## 竞态测试 (Race Detection)

> 所有竞态测试运行方式: `go test -race -count=N ./service/`
>
> **注意:** Windows 下 `-race` 需要 CGO (CGO_ENABLED=1) 和 GCC 工具链。
> Linux/macOS 下可直接运行 `go test -race`。

### 异步 Worker Pool 竞态场景

| ID | Go 测试函数 | Goroutine 分布 | 持续时间 | 检测目标 |
|----|-----------|---------------|---------|---------|
| RT-001 | `TestRace_ConcurrentSubmit` | 100 goroutines × 10 submits | ~5s | asyncStarted 读/写, channel send |
| RT-002 | `TestRace_StartShutdownRace` | 20 并发 start + 20 并发 shutdown | ~3s | asyncMu 锁, close(asyncShutdown) |
| RT-003 | `TestRace_ShutdownWhileWriting` | 4 workers 工作中 → shutdown | ~3s | select 分支竞争 |
| RT-004 | `TestRace_HotLoop` | 持续读写 + 定期启停 | 5s (short=30s) | 长期运行的 race 累积 |
| RT-005 | `TestRace_MultiServiceManager` | 5 个 sm 实例并发操作 | ~5s | 全局变量读安全 |

### 缓存操作竞态场景

| ID | Go 测试函数 | 场景 | 检测目标 |
|----|-----------|------|---------|
| RT-101 | `TestRace_WriteReadSameKey` | 10 writers + 20 readers on same key | Pipeline 安全 |
| RT-102 | `TestRace_InvalidateWhileReading` | LookupQuery + InvalidateCache 并发 | 结果一致性 |
| RT-103 | `TestRace_VersionWriteRace` | 5 goroutines 并发 WritedownSingleWithVersion | Watch/TxPipelined 原子性 |
| RT-104 | `TestRace_PatternOpsRace` | ScanKeys + InvalidateCacheByPattern 并发 | SCAN 游标安全 |

### 注册表竞态

| ID | Go 测试函数 | 检测目标 |
|----|-----------|---------|
| RT-201 | `TestRace_ConcurrentTranslate` | GORM registry.translators map 并发读 |
| RT-202 | `TestRace_ConcurrentTranslate_Redis` | Redis registry.translators map 并发读 |
| RT-301 | `TestRace_ConcurrentSetGet` | go-redis 客户端连接池并发安全性 |

## 性能基准测试

### 基准测试运行方式

```bash
go test -bench=. -benchmem -benchtime=3s ./pkg/...
go test -bench=BenchmarkWritedownQuery -benchmem -count=5 ./service/ > results.txt
```

### 缓存写入性能

| ID | Benchmark 名称 | 数据量 | 关键指标 |
|----|---------------|--------|---------|
| BM-001 | `BenchmarkWritedownQuery_Pipeline_100` | 100 rows | ns/op, B/op, allocs/op |
| BM-002 | `BenchmarkWritedownQuery_Pipeline_1000` | 1000 rows | throughput (rows/s) |
| BM-003 | `BenchmarkWritedownQuery_Pipeline_10000` | 10000 rows | throughput + memory |
| BM-004 | `BenchmarkWritedownQuery_BatchSize_50` | 1000 rows, bs=50 | 对比最佳 batch |
| BM-005 | `BenchmarkWritedownQuery_BatchSize_100` | 1000 rows, bs=100 | 对比最佳 batch |
| BM-006 | `BenchmarkWritedownQuery_BatchSize_500` | 1000 rows, bs=500 | 对比最佳 batch |
| BM-007 | `BenchmarkWritedownSingle_Set` | 1 row | 单次写入延迟 |
| — | `BenchmarkWritedownQuery_Baseline` | 1000 rows | baseline 对照 |
| — | `BenchmarkSetMultiple_100Items` | 100 items | Pipeline 批量 Set |
| — | `BenchmarkGetMultiple_100Items` | 100 items | Pipeline 批量 Get |

### 缓存读取性能

| ID | Benchmark | 场景 | 期望延迟 |
|----|----------|------|---------|
| BM-101 | `BenchmarkLookupQuery_10Keys` | 10 keys 全命中 | < 1ms |
| BM-102 | `BenchmarkLookupQuery_100Keys` | 100 keys 全命中 | < 5ms |
| BM-103 | `BenchmarkLookupSingleWithFallback_Hit` | 缓存命中 (redigo Get+Scan) | < 100µs |
| BM-104 | `BenchmarkLookupSingleWithFallback_Miss` | 缓存miss (redigo Nil 路径) | < 1ms |

### 序列化性能

| ID | Benchmark | 描述 |
|----|----------|------|
| BM-201 | `BenchmarkMarshalForRedis` | json.Marshal 单对象 |
| BM-202 | `BenchmarkUnmarshalForRedis` | json.Unmarshal 单对象 |
| BM-203 | `BenchmarkExtractID_JSON` | JSON 往返提取 ID |
| — | `BenchmarkBuildCacheKey_Uint` | buildCacheKey(uint) |
| — | `BenchmarkBuildCacheKey_String` | buildCacheKey(string) |
| — | `BenchmarkExtractIDFromKey` | extractIDFromKey 解析 |
| — | `BenchmarkNewServiceManager` | NewServiceManager 创建开销 |

### 过滤性能

| ID | Benchmark | 描述 |
|----|----------|------|
| BM-301 | `BenchmarkRedisFilter_1Filter_100Keys` | 单过滤器+100keys |
| BM-302 | `BenchmarkRedisFilter_5Filters_1000Keys` | 5过滤器链+1000keys |
| BM-303 | `BenchmarkGormFilter_10Filters` | 10个过滤器 ApplyGorm (纯 filter 构建) |

### 并发吞吐量

| ID | Benchmark | 读写比 | 并发数 | 期望吞吐 |
|----|----------|--------|--------|---------|
| BM-401 | `BenchmarkThroughput_ReadHeavy` | 80R/20W | GOMAXPROCS | ≥ 5000 ops/s |
| BM-402 | `BenchmarkThroughput_WriteHeavy` | 20R/80W | GOMAXPROCS | ≥ 2000 ops/s |
| BM-403 | `BenchmarkThroughput_Mixed` | 50R/50W | GOMAXPROCS | ≥ 3000 ops/s |

### 辅助基准测试

| Benchmark | 描述 |
|-----------|------|
| `BenchmarkScanKeys_1000Keys` | SCAN 遍历 1000 keys |

### 性能回归阈值

| 指标 | 允许波动 | 严重退化 |
|------|---------|---------|
| 缓存写入延迟 (P50) | ±10% | > +30% |
| 缓存读取延迟 (P50) | ±5% | > +20% |
| 内存分配 | ±15% | > +50% |
| 并发吞吐 | ±10% | > -25% |

---

## 实现详情

### 文件结构

| 文件 | 内容 |
|------|------|
| [tests/race_and_perf_test.go](../../tests/race_and_perf_test.go) | 全部竞态测试 + 全部性能基准测试（外部测试包 `package tests`） |

### 架构说明

测试文件位于 `tests/` 目录下，作为外部测试包（`package tests`）：
- 通过 `TestMain` 启动 miniredis 并通过环境变量注入 `service.InitRedis()` 完成全局初始化
- 使用公开 API（`service.ServiceManager`、`service.GetRedis()` 等），不直接访问 unexported 成员
- 每个测试通过 `flushRedis()` 清空 miniredis 确保测试隔离

### 竞态测试运行命令

```bash
# 全部竞态测试 (Linux/macOS)
go test -race -run="TestRace_" -count=3 -timeout 120s ./tests/

# 全部竞态测试 (Windows — 需 CGO 和 GCC)
CGO_ENABLED=1 go test -race -run="TestRace_" -count=3 -timeout 120s ./tests/

# 单个竞态测试
go test -race -run=TestRace_ConcurrentSubmit -count=5 ./tests/

# 跳过长时间测试
go test -race -run="TestRace_" -short -count=1 ./tests/

# 异步 Worker Pool 竞态
go test -race -run="TestRace_ConcurrentSubmit|TestRace_StartShutdownRace|TestRace_ShutdownWhileWriting|TestRace_HotLoop|TestRace_MultiServiceManager" -count=3 ./tests/

# 缓存操作竞态
go test -race -run="TestRace_WriteReadSameKey|TestRace_InvalidateWhileReading|TestRace_VersionWriteRace|TestRace_PatternOpsRace" -count=5 ./tests/

# 注册表竞态
go test -race -run="TestRace_ConcurrentTranslate|TestRace_ConcurrentTranslate_Redis" -count=5 ./tests/
```

### 性能基准测试运行命令

```bash
# 全部基准测试 (跳过单元测试)
go test -bench=. -benchmem -benchtime=3s -run="^$" ./tests/

# 缓存写入性能
go test -bench="BenchmarkWritedown" -benchmem -benchtime=3s -run="^$" -count=5 ./tests/ > bench_write_results.txt

# 缓存读取性能
go test -bench="BenchmarkLookup|BenchmarkLookupSingleWithFallback" -benchmem -benchtime=3s -run="^$" ./tests/

# 序列化性能
go test -bench="BenchmarkMarshal|BenchmarkUnmarshal|BenchmarkExtractID" -benchmem -benchtime=3s -run="^$" ./tests/

# 过滤性能
go test -bench="BenchmarkRedisFilter|BenchmarkGormFilter" -benchmem -benchtime=3s -run="^$" ./tests/

# 并发吞吐量
go test -bench="BenchmarkThroughput" -benchmem -benchtime=3s -run="^$" ./tests/

# 对比不同 batch size 的写入性能
go test -bench="BenchmarkWritedownQuery_BatchSize" -benchmem -benchtime=1s -run="^$" ./tests/
```

### 测试配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 测试 Redis | miniredis (内存) | 所有竞态/基准测试使用 miniredis 模拟，无需真实 Redis |
| 测试 DB | 无需 DB | 竞态测试仅覆盖缓存层和 worker pool 层 |
| HotLoop 持续时间 | 5 秒 | 通过 `testing.Short()` 在 CI 中可跳过 |
| 基准测试时间 | 3 秒 (benchtime) | 可通过 `-benchtime` 调整 |

### 已知问题

| 问题 | 影响测试 | 严重度 | 状态 |
|------|---------|--------|------|
| `ShutdownAsyncWorkers` close of closed channel | RT-002 | 中 | 测试已使用 recover 捕获，待修复源代码 |
| task queue full (1000 tasks → 256 queue) | RT-001 | 低 | 预期行为，验证了背压机制 |
