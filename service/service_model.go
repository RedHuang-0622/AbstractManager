package service

import (
	"context"
	"reflect"
	"sync"
	"time"
)

// asyncCacheTask 异步缓存写入任务
type asyncCacheTask[T any] struct {
	ctx        context.Context
	key        string
	data       *T
	expiration time.Duration
}

type ServiceManager[T any] struct {
	Resource     T      // 被管理的资源
	ResourceName string // 资源名称
	TableName    string // 表名
	Schema       string // 数据库模式
	CacheKeyType string // 缓存键
	CacheKeyName string // 缓存键名称

	// 异步写入 worker pool
	asyncTasks    chan asyncCacheTask[T] // 任务队列
	asyncWg       sync.WaitGroup         // 等待所有 worker 完成
	asyncShutdown chan struct{}          // 通知 worker 退出
	asyncStarted  bool                   // 是否已启动 worker
	asyncMu       sync.Mutex             // 保护 asyncStarted
}

func getTypeName[T any](value T) string {
	var zero T
	t := reflect.TypeOf(zero)
	// 如果 T 是指针，剥一层
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// NewServiceManager 创建一个新的 ServiceManager 实例
// 通过reflect获取名字自动赋值给ResourceName和TableName还有keyname
func NewServiceManager[T any](resource T) *ServiceManager[T] {
	return &ServiceManager[T]{
		Resource:      resource,
		ResourceName:  getTypeName(resource),
		TableName:     getTypeName(resource),
		Schema:        "public",
		CacheKeyType:  "none",
		CacheKeyName:  getTypeName(resource) + "_key",
		asyncTasks:    make(chan asyncCacheTask[T], 256),
		asyncShutdown: make(chan struct{}),
	}
}
