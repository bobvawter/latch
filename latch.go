// Copyright 2020 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package latch provides a notification-based counter latch.
package latch

import (
	"errors"
	"sync"
	"sync/atomic"
)

// https://github.com/golang/go/issues/8005#issuecomment-190753527
type noCopy struct{}

func (*noCopy) Lock() {}

type waiter struct {
	addHolds    int               // Must be >= 0
	notify      chan<- WaitStatus // Capacity 1
	leaveLocked bool
}

// Counter is conceptually similar to a sync.WaitGroup, except that it
// does not require foreknowledge of how many tasks there will be, and
// it uses a notification-style API. That it, it can count both up and
// down.
//
// It is always valid to add or remove holds on the Counter, regardless
// of their ordering with Wait. For example, a Counter can be used to
// track the completion of tasks that spawn an unknown number of
// sub-tasks.
//
// The Wait methods on a Counter will trigger when the count
// is exactly zero.
//
// Counter additionally implements sync.Locker.  When the latch is in a
// locked state, no changes to the Counter's count may be made.
//
// A Counter should not be copied.
type Counter struct {
	atomic struct {
		count int64
	}
	mu struct {
		sync.Locker
		waiters []*waiter
	}
	noCopy noCopy
}

// New constructs a Counter using the provided options.
func New(options ...*Option) *Counter {
	ret := &Counter{}
	for _, opt := range options {
		if opt.locker != nil {
			ret.mu.Locker = opt.locker
		}
	}
	if ret.mu.Locker == nil {
		ret.mu.Locker = &sync.Mutex{}
	}
	return ret
}

// Apply changes the hold-count on the latch.
//
// This method will panic if a negative delta exceeds the number of
// holds on the Counter.
func (l *Counter) Apply(delta int) {
	l.Lock()
	l.applyLocked(delta)
	l.Unlock()
}

// Count returns an estimate of the number of pending holds.
//
// This method is not subject to locking and is safe to call at any
// time. This makes it suitable as a source for metrics collection.
func (l *Counter) Count() int64 {
	return atomic.LoadInt64(&l.atomic.count)
}

// Hold increments the use-count by 1.
//
// It is a shortcut for Apply(1).
func (l *Counter) Hold() {
	l.Apply(1)
}

// Lock will lock the underlying sync.Locker used by the Counter.
//
// This has the effect of blocking all other uses of the Counter
// until a call to Unlock is made.
func (l *Counter) Lock() {
	l.mu.Lock()
}

// Release decrements the use-count by 1.
//
// It is a shortcut for Apply(-1).
func (l *Counter) Release() {
	l.Apply(-1)
}

// Unlock will unlock the underlying sync.Locker used by the Counter.
func (l *Counter) Unlock() {
	if l.maybeNotifyLocked() {
		l.mu.Unlock()
	}
}

// Wait returns a channel that will emit a single value once the wait
// condition has been (instantaneously) satisfied.
//
// Consider using WaitLock if it is necessary to prevent additional
// holds from being obtained until a call to Unlock is made.
func (l *Counter) Wait() <-chan WaitStatus {
	return l.wait(false, 0)
}

// WaitHold returns a channel that will emit a single value once the
// wait condition has been (instantaneously) satisfied and the requested
// number of holds has been added.
//
// This method will panic if holds is negative.
func (l *Counter) WaitHold(holds int) <-chan WaitStatus {
	if holds < 0 {
		panic(errors.New("holds must be >= 0"))
	}

	return l.wait(false, holds)
}

// WaitLock returns a channel that will emit a single value once the
// wait condition has been (instantaneously) satisfied and the Counter
// has been left in a locked state.
//
// Callers to this method must ensure that Unlock is called after
// receiving the notification.
func (l *Counter) WaitLock() <-chan WaitStatus {
	return l.wait(true, 0)
}

func (l *Counter) applyLocked(delta int) {
	res := atomic.AddInt64(&l.atomic.count, int64(delta))
	if res < 0 {
		panic(errors.New("latch was over-released"))
	}
}

// maybeNotifyLocked will drain waiters until it encounters a
// non-trivial waiter. A non-trivial waiter is one with a non-zero
// delta, or a request to leave the Counter in a locked state.
//
// This method will return true if the Locker should be unlocked.
func (l *Counter) maybeNotifyLocked() (unlock bool) {
	unlock = true

	holds := atomic.LoadInt64(&l.atomic.count)
	if holds != 0 {
		return
	}

	for len(l.mu.waiters) > 0 {
		earlyExit := false
		status := delayed

		// Dequeue.
		wake := l.mu.waiters[0]
		l.mu.waiters[0] = nil
		l.mu.waiters = l.mu.waiters[1:]

		if wake.addHolds > 0 {
			atomic.AddInt64(&l.atomic.count, int64(wake.addHolds))
			earlyExit = true
		}
		if wake.leaveLocked {
			earlyExit = true
			status = status.locked()
			unlock = false
		}

		wake.notify <- status
		close(wake.notify)

		if earlyExit {
			return
		}
	}
	return
}

func (l *Counter) wait(leaveLocked bool, delta int) <-chan WaitStatus {
	ch := make(chan WaitStatus, 1)

	// Fast-path: Avoid locking if the count is currently 0.
	if !leaveLocked && atomic.CompareAndSwapInt64(&l.atomic.count, 0, int64(delta)) {
		ch <- 0
		close(ch)
		return ch
	}

	w := &waiter{
		addHolds:    delta,
		notify:      ch,
		leaveLocked: leaveLocked,
	}

	l.Lock()
	// Slower path: Immediately notify if the count is currently 0.
	if leaveLocked && atomic.CompareAndSwapInt64(&l.atomic.count, 0, int64(delta)) {
		ch <- locked
		close(ch)
		return ch
	}
	l.mu.waiters = append(l.mu.waiters, w)
	l.Unlock()

	return ch
}
