package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type MergeLockPayload struct {
	Status    string `json:"status"` // e.g. MERGING
	Prefix    string `json:"prefix"`
	MergeID   int64  `json:"merge_id"`   // business merge id
	StartedAt int64  `json:"started_at"` // unix timestamp
}

type MergeLockService struct {
	redis redis.Cmdable
	key   string
}

func NewMergeLockService(key, address, port, password string, db int) *MergeLockService {
	client := redis.NewClient(&redis.Options{
		Addr:     address + ":" + port,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	return &MergeLockService{
		redis: client,
		key:   key,
	}
}

func (s *MergeLockService) Acquire(ctx context.Context, mergeID int64, prefix string) error {
	payload := MergeLockPayload{
		Status:    "MERGING",
		Prefix:    prefix,
		MergeID:   mergeID,
		StartedAt: time.Now().Unix(),
	}

	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ok, err := s.redis.SetNX(ctx, s.key, value, 0).Result() // NO TTL
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("apply approval flow already in progress")
	}

	return nil
}

func (s *MergeLockService) Release(ctx context.Context) error {
	return s.redis.Del(ctx, s.key).Err()
}

func (s *MergeLockService) IsLocked(ctx context.Context) (bool, error) {
	exists, err := s.redis.Exists(ctx, s.key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (s *MergeLockService) Get(ctx context.Context) (*MergeLockPayload, error) {
	val, err := s.redis.Get(ctx, s.key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var payload MergeLockPayload
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
