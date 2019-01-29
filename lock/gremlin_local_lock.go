package lock

import (
	"sync"
	"time"

	c "github.com/patrickmn/go-cache"
)

// The LocalLock is an implementation of the LockClient_i/Lock_i structure
// That is designed to be used in a non-distributed environment
// It uses a cache of Mutexes which are locked and unlocked according to a specified key
type LocalLockClient struct {
	Keys           *c.Cache
	ExpirationTime time.Duration
}

func NewLocalLockClient(expirationTime, purgeTime int) *LocalLockClient {
	c := c.New(time.Duration(expirationTime)*time.Minute, time.Duration(purgeTime)*time.Minute)
	return &LocalLockClient{
		Keys:           c,
		ExpirationTime: time.Duration(expirationTime),
	}
}

func (c *LocalLockClient) LockKey(key string) (Lock_i, error) {
	_, found := c.Keys.Get(key)
	if !found {
		c.Keys.SetDefault(key, &sync.Mutex{})
	}
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
	var m *sync.Mutex
	val, found := lock.Client.Keys.Get(lock.Key)
	if !found {
		m = &sync.Mutex{}
		err := lock.Client.Keys.Add(lock.Key, m, lock.Client.ExpirationTime)
		if err != nil {
			return err
		}
	} else {
		m = val.(*sync.Mutex)
	}
	m.Lock()
	return nil
}

func (lock LocalLock) Unlock() error {

	var m *sync.Mutex
	val, found := lock.Client.Keys.Get(lock.Key)
	if !found {
		m = &sync.Mutex{}
		err := lock.Client.Keys.Add(lock.Key, m, lock.Client.ExpirationTime)
		if err != nil {
			return err
		}
	} else {
		m = val.(*sync.Mutex)
	}
	m.Unlock()
	return nil
}
