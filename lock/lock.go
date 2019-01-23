package lock

// LockClient_i is an interface for concurrency locking mechanisms
// With a set of keys which can be locked and unlocked independently of each other
// LockClient_i requires a LockKey method, which accepts a string key and returns the corresponding Lock for that key

type LockClient_i interface {
	LockKey(key string) (Lock_i, error)
}

// Lock_i is an instance of a Lock in the set of LockClient_i locks
// It requires 3 methods for implementation
//    - Lock: which attempts to establish a lock on its key
//    - Unlock: which attempts to remove a lock on its key
//    - Destroy: which deletes the Lock_i's key from its parent client
type Lock_i interface {
	Lock() error
	Unlock() error
	Destroy() error
}
