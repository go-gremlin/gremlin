package lock

import (
	b "encoding/base64"
	"fmt"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

// The ConsulLock is an implementation of the LockClient_i/Lock_i structure
// That is designed to be used in a distributed environment
// It creates a Consul API client, which uses its native Lock struct to create & lock/unlock KVs

type ConsulAPIWrapper struct {
	DefaultConfig func() *consulapi.Config
	NewClient     func(config *consulapi.Config) (*consulapi.Client, error)
}

type ConsulLockClient struct {
	Client      *consulapi.Client
	LockOptions *consulapi.LockOptions
	BaseFolder  string
}

type ConsulLock struct {
	Key        string
	ConsulLock *consulapi.Lock
	LockCh     <-chan struct{}
}

func NewConsulLockClient(address, baseFolder string, lockOptions *consulapi.LockOptions) (ConsulLockClient, error) {
	consulAPI := ConsulAPIWrapper{
		DefaultConfig: consulapi.DefaultConfig,
		NewClient:     consulapi.NewClient,
	}
	return newConsulLockClient(address, baseFolder, consulAPI, lockOptions)
}

func newConsulLockClient(address, baseFolder string, consulAPI ConsulAPIWrapper, lockOptions *consulapi.LockOptions) (ConsulLockClient, error) {
	var (
		err          error
		lock         ConsulLockClient
		consulConfig *consulapi.Config
		consulClient *consulapi.Client
	)

	if consulConfig = consulAPI.DefaultConfig(); consulConfig == nil {
		return lock, fmt.Errorf("Unable to get consul config")
	}

	consulConfig.Address = address

	if consulClient, err = consulAPI.NewClient(consulConfig); err != nil {
		return lock, err
	}

	if consulClient == nil {
		return lock, fmt.Errorf("Unable to get consul config")
	}
	lock.Client = consulClient

	lock.LockOptions = lockOptions
	lock.BaseFolder = baseFolder
	return lock, nil
}

func (c ConsulLockClient) LockKey(key string) (Lock_i, error) {
	lockKeyHash := b.StdEncoding.EncodeToString([]byte(key))
	lockOptions := c.LockOptions
	lockOptions.Key = makeFullKey(c.BaseFolder, lockKeyHash)
	lockOpts, err := c.Client.LockOpts(lockOptions)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Consul lock for key %s with error: %v", key, err)
	}

	lock := ConsulLock{
		Key:        lockKeyHash,
		ConsulLock: lockOpts,
	}
	return lock, nil
}

func makeFullKey(baseFolder, key string) string {
	fullKey := key
	if baseFolder != "" {
		fullKey = fmt.Sprintf("%s/%s", strings.TrimRight(baseFolder, "/"), key)
	}
	return fullKey
}

func (lock ConsulLock) Lock() error {
	stopCh := make(chan struct{})
	lockCh, err := lock.ConsulLock.Lock(stopCh)
	if err != nil {
		return fmt.Errorf("Error locking key %s: %v", lock.Key, err)
	}
	lock.LockCh = lockCh
	return nil
}

func (lock ConsulLock) Unlock() error {
	err := lock.ConsulLock.Unlock()
	if err != nil {
		return err
	}
	lock.ConsulLock.Destroy()
	return nil
}
