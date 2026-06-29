package hooks

import (
	// "fmt"
	"context"
	"log"
	"sync"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type HookStage string

const (
	BeforeInsert HookStage = "before_insert"
	BeforeUpdate HookStage = "before_update"
	BeforeSave   HookStage = "before_save"
	BeforeDelete HookStage = "before_delete"
	AfterInsert  HookStage = "after_insert"
	AfterUpdate  HookStage = "after_update"
	AfterDelete  HookStage = "after_delete"
	AfterSave    HookStage = "after_save"
	AfterApprove HookStage = "after_approve"
)

type HookFunc func(ctx context.Context, data any, prefix string) types.Message

type hookRegistry struct {
	mu sync.RWMutex

	hooks map[string]map[HookStage][]HookFunc

	warningThreshold  time.Duration
	criticalThreshold time.Duration

	// callback khi xảy ra critical alert (optional)
	criticalAlertCallback func(model string, stage HookStage, duration time.Duration)
}

var globalRegistry = &hookRegistry{
	hooks:             make(map[string]map[HookStage][]HookFunc),
	warningThreshold:  300 * time.Millisecond,
	criticalThreshold: 1000 * time.Millisecond,
}

func DefaultConfig() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.warningThreshold = 300 * time.Millisecond
	globalRegistry.criticalThreshold = 500 * time.Millisecond
}

// Configure thresholds và callback cảnh báo
func Configure(warning, critical time.Duration, criticalCallback func(model string, stage HookStage, duration time.Duration)) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.warningThreshold = warning
	globalRegistry.criticalThreshold = critical
	globalRegistry.criticalAlertCallback = criticalCallback
}

func wrapHook(modelName string, stage HookStage, fn HookFunc) HookFunc {
	return func(ctx context.Context, data any, prefix string) types.Message {
		start := time.Now()
		err := fn(ctx, data, prefix)
		elapsed := time.Since(start)

		log.Printf("[hook] model=%s stage=%s duration=%s error=%v", modelName, stage, elapsed, err.MessageErr)

		if elapsed > globalRegistry.criticalThreshold {
			log.Printf("[hook CRITICAL] model=%s stage=%s duration=%s", modelName, stage, elapsed)
			if globalRegistry.criticalAlertCallback != nil {
				go globalRegistry.criticalAlertCallback(modelName, stage, elapsed)
			}
		} else if elapsed > globalRegistry.warningThreshold {
			log.Printf("[hook WARNING] model=%s stage=%s duration=%s", modelName, stage, elapsed)
		}

		return err
	}
}

func Register[T any](stage HookStage, fn HookFunc) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	modelName := helper.GetNameModel[T]()
	if globalRegistry.hooks[modelName] == nil {
		globalRegistry.hooks[modelName] = make(map[HookStage][]HookFunc)
	}

	globalRegistry.hooks[modelName][stage] = append(
		globalRegistry.hooks[modelName][stage],
		wrapHook(modelName, stage, fn),
	)
}

func Run(ctx context.Context, modelName, prefix string, stage HookStage, data any) types.Message {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	for _, fn := range globalRegistry.hooks[modelName][stage] {
		if err := fn(ctx, data, prefix); err.HasError() {
			// return fmt.Errorf("hook failed on %s:%s: %w", modelName, stage, err)
			return err
		}
	}
	return types.Message{}
}
