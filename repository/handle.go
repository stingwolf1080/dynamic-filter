package repository

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type HandleFunc func(ctx context.Context, data any, prefix string) types.Message

type hookRegistry struct {
	mu    sync.RWMutex
	hooks map[string]map[string]HandleFunc
}

var globalRegistry = &hookRegistry{
	hooks: make(map[string]map[string]HandleFunc),
}

func wrapHook(modelName string, stage string, fn HandleFunc) HandleFunc {
	return func(ctx context.Context, data any, prefix string) types.Message {
		start := time.Now()
		err := fn(ctx, data, prefix)
		elapsed := time.Since(start)
		log.Printf("[repository] model=%s stage=%s duration=%s error=%v\n", modelName, stage, elapsed, err.MessageErr)
		return err
	}
}

func registerHandle(modelName, stage string, fn HandleFunc) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	if globalRegistry.hooks[modelName] == nil {
		globalRegistry.hooks[modelName] = make(map[string]HandleFunc)
	}
	globalRegistry.hooks[modelName][stage] = wrapHook(modelName, stage, fn)
}

func callFunc(ctx context.Context, modelName, prefix, stage string, data any) types.Message {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	if _, ok := globalRegistry.hooks[modelName][stage]; !ok {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Data error, please contact adminstrator",
			MessageErr: fmt.Errorf("Customer handle repository function %s not found", stage),
		}
	}
	fn := globalRegistry.hooks[modelName][stage]
	return fn(ctx, data, prefix)
}

var globalDeleteFunc = &hookDeleteRegistry{
	hooks: make(map[string]GenericDeleter),
}

type hookDeleteRegistry struct {
	mu    sync.RWMutex
	hooks map[string]GenericDeleter
}

func RegisterDeleteFunc[T any](name string, r Repository[T]) {
	globalDeleteFunc.mu.Lock()
	globalDeleteFunc.hooks[name] = r
	globalDeleteFunc.mu.Unlock()
}

func ResolveRepo(name string) (GenericDeleter, bool) {
	globalDeleteFunc.mu.RLock()
	r, ok := globalDeleteFunc.hooks[name]
	globalDeleteFunc.mu.RUnlock()
	return r, ok
}
