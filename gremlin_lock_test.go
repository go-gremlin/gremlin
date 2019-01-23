package gremlin

import (
	"fmt"
	"testing"
	"time"
)

func TestLockClientLockKey(t *testing.T) {
	c := NewLocalLockClient()
	key := "testKey"
	lockI, err := c.LockKey(key)
	lock := lockI.(LocalLock)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create Lock with following error: %s", err))
		return
	}
	_, ok := c.Keys.Load(key)
	if !ok {
		t.Error("testKey was not created in Client map")
		return
	}
	if lock.Key != key || lock.LockWaitTime != DEFAULT_LOCK_WAIT_TIME || lock.MaxRetries != DEFAULT_MAX_RETRIES {
		t.Error("Config error in lock: ", lock)
		return
	}
}

func TestClientLockWorks(t *testing.T) {
	c := NewLocalLockClient()
	key := "testKey"
	var vals []int

	lock, err := c.LockKey(key)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create Lock with following error: %s", err))
		return
	}
	lock2, err := c.LockKey(key)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create Lock with following error: %s", err))
		return
	}

	go func() {
		lock.Lock()
		time.Sleep(2 * time.Second)
		vals = append(vals, 0)
		lock.Unlock()
	}()

	time.Sleep(1 * time.Second)

	lock2.Lock()
	vals = append(vals, 1)
	lock2.Unlock()

	if vals[0] != 0 || vals[1] != 1 {
		t.Error("Your mutex isn't working!")
	}
}

func TestClientLockWorksAdvanced(t *testing.T) {
	c := NewLocalLockClient()
	key := "testKey"
	key3 := "testKey3"
	var vals []int

	lock, err := c.LockKey(key)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create Lock with following error: %s", err))
		return
	}
	lock2, err := c.LockKey(key)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create Lock with following error: %s", err))
		return
	}
	lock3, err := c.LockKey(key3)
	if err != nil {
		t.Error(fmt.Sprintf("Failed to create Lock with following error: %s", err))
		return
	}

	go func() {
		lock.Lock()
		time.Sleep(2 * time.Second)
		vals = append(vals, 0)
		lock.Unlock()
	}()

	time.Sleep(1 * time.Second)

	lock3.Lock()
	vals = append(vals, 2)
	lock3.Unlock()

	lock2.Lock()
	vals = append(vals, 1)
	lock2.Unlock()

	if vals[0] != 2 || vals[1] != 0 || vals[2] != 1 {
		t.Error("Your mutex isn't working!")
	}
}
