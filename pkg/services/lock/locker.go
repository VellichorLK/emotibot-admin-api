/*
Package lock contains the locking process.
It only describe the behavior of the algorithm but not implementations.
*/
package lock

// Locker Describe a locking algorithm which should provide liveness of the server by the Lock's notify channel response.
// It also should be cancelable by Sending a signal into stopCh or calling Unlock of the Locker.
type Locker interface {
	//	Lock attempt to acquire lock, if acquired it return notify to notify subscriber if the lock have been released by server.
	Lock(stopCh <-chan struct{}) (notify <-chan struct{}, err error)
	// 	Unlock will released the lock. return ErrLockNotExist if the lock is not currently held or non-nil error if release had failed.
	Unlock() error
}
