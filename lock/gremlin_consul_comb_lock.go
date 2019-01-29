package lock

import (
	consulapi "github.com/hashicorp/consul/api"
)

type ConsulCombinationLockClient struct {
	ConsulClient *ConsulLockClient
	LocalClient  *LocalLockClient
}

type ConsulCombinationLock struct {
	ConsulLock ConsulLock
	LocalLock  LocalLock
}

func NewConsulCombinationLockClient(consulAddress, consulBaseFolder string, consulLockOptions *consulapi.LockOptions, localExpirationTime, localPurgeTime int) (*ConsulCombinationLockClient, error) {
	consulClient, err := NewConsulLockClient(consulAddress, consulBaseFolder, consulLockOptions)
	if err != nil {
		return nil, err
	}
	localClient := NewLocalLockClient(localExpirationTime, localPurgeTime)
	return &ConsulCombinationLockClient{
		ConsulClient: &consulClient,
		LocalClient:  localClient,
	}, nil
}

func (c ConsulCombinationLockClient) LockKey(key string) (Lock_i, error) {
	consulLock, err := c.ConsulClient.LockKey(key)
	if err != nil {
		return nil, err
	}
	localLock, err := c.LocalClient.LockKey(key)
	if err != nil {
		return nil, err
	}
	return ConsulCombinationLock{
		ConsulLock: consulLock.(ConsulLock),
		LocalLock:  localLock.(LocalLock),
	}, nil
}

func (lock ConsulCombinationLock) Lock() error {
	err := lock.LocalLock.Lock()
	if err != nil {
		return err
	}
	err = lock.ConsulLock.Lock()
	return err
}

func (lock ConsulCombinationLock) Unlock() error {
	err := lock.ConsulLock.Unlock()
	if err != nil {
		return err
	}
	err = lock.LocalLock.Unlock()
	return err
}
