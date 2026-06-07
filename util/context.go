package util

import (
	"context"
	"os"
	"strconv"
	"time"
)

// GetDefaultDBTimeout 从环境变量 DB_TIMEOUT_SECONDS 读取默认数据库操作超时时间
// 默认 30 秒
func GetDefaultDBTimeout() time.Duration {
	if s := os.Getenv("DB_TIMEOUT_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return time.Duration(v) * time.Second
		}
	}
	return 30 * time.Second
}

// GetDefaultRedisTimeout 从环境变量 REDIS_TIMEOUT_SECONDS 读取默认 Redis 操作超时时间
// 默认 10 秒
func GetDefaultRedisTimeout() time.Duration {
	if s := os.Getenv("REDIS_TIMEOUT_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return time.Duration(v) * time.Second
		}
	}
	return 10 * time.Second
}

// GetDefaultDDLTimeout 从环境变量 DDL_TIMEOUT_SECONDS 读取默认 DDL 操作超时时间
// 默认 60 秒（DDL 操作通常较慢）
func GetDefaultDDLTimeout() time.Duration {
	if s := os.Getenv("DDL_TIMEOUT_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return time.Duration(v) * time.Second
		}
	}
	return 60 * time.Second
}

// EnsureTimeout 如果 ctx 没有 deadline，则添加 defaultTimeout 作为兜底超时。
// 如果 ctx 已有 deadline，直接返回原 ctx（不覆盖调用方的超时设置）。
//
// 使用方式：
//
//	ctx, cancel := util.EnsureTimeout(ctx, util.GetDefaultDBTimeout())
//	defer cancel()
//	db := GetDB().WithContext(ctx)
func EnsureTimeout(ctx context.Context, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, defaultTimeout)
}
