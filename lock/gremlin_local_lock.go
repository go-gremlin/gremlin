package lock

import (
	"fmt"
	"sync"
)

// The LocalLock is an implementation of the LockClient_i/Lock_i structure
// That is designed to be used in a non-distributed environment
// It uses a sync Map of Mutexes which are locked and unlocked according to a specified key
type LocalLockClient struct {
	Keys *sync.Map
}

func NewLocalLockClient() *LocalLockClient {
	keyMap := sync.Map{}
	return &LocalLockClient{
		Keys: &keyMap,
	}
}

func (c *LocalLockClient) LockKey(key string) (Lock_i, error) {
	_, _ = c.Keys.LoadOrStore(key, &sync.Mutex{})
	lock := LocalLock{
		Client: c,
		Key:    key,
	}
	return lock, nil
}

type LocalLock struct {
	Client *LocalLockClient
	Key    string
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
	_, _ = lock.Client.Keys.LoadOrStore(lock.Key, lockVal)
	return nil
}

func (lock LocalLock) Destroy() error {
	lock.Client.Keys.Delete(lock.Key)
	return nil
}
