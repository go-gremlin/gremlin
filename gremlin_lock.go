package gremlin

import (
	"fmt"
	"sync"
	"time"
)

type LockClient_i interface {
	LockKey(key string) (Lock_i, error)
}

type Lock_i interface {
	Lock() error
	Unlock() error
	Destroy() error
}

// Creates a lock using a sync map
type LocalLockClient struct {
	Keys sync.Map
}

func NewLocalLockClient() *LocalLockClient {
	keyMap := sync.Map{}
	return &LocalLockClient{
		Keys: keyMap,
	}
}

func (c *LocalLockClient) LockKey(key string) (Lock_i, error) {
	_, _ = c.Keys.LoadOrStore(key, &sync.Mutex{})
	lock := LocalLock{
		Client:       c,
		Key:          key,
		LockWaitTime: DEFAULT_LOCK_WAIT_TIME,
		MaxRetries:   DEFAULT_MAX_RETRIES,
	}
	return lock, nil
}

type LocalLock struct {
	Client       *LocalLockClient
	Key          string
	LockWaitTime time.Duration
	MaxRetries   int
}

func (lock LocalLock) Lock() error {

	lockVal, ok := lock.Client.Keys.Load(lock.Key)
	lockMutex := lockVal.(*sync.Mutex)
	if !ok || lockMutex == nil {
		return fmt.Errorf("Error loading key %s from lock.", lock.Key)
	}
	lockMutex.Lock()
	return nil
}

func (lock LocalLock) Unlock() error {
	lockVal, ok := lock.Client.Keys.Load(lock.Key)
	lockMutex := lockVal.(*sync.Mutex)
	if !ok || lockMutex == nil {
		return fmt.Errorf("Error loading key %s from lock.", lock.Key)
	}
	lockMutex.Unlock()
	return nil
}

func (lock LocalLock) Destroy() error {
	lock.Client.Keys.Delete(lock.Key)
	return nil
}
