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

package latch

import (
	"testing"

	"sync"

	"sync/atomic"

	"github.com/stretchr/testify/assert"
)

func Benchmark(b *testing.B) {
	b.Run("serial", func(b *testing.B) {
		l := New()
		for i := 0; i < b.N; i++ {
			l.Hold()
			l.Release()
		}
	})
	b.Run("parallel", func(b *testing.B) {
		l := New()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Hold()
				l.Release()
			}
		})
	})
}

func TestWait(t *testing.T) {
	a := assert.New(t)

	l := New()
	a.Equal(int64(0), l.Count())
	a.False((<-l.Wait()).Delayed())

	l.Hold()
	a.Equal(int64(1), l.Count())

	ch := l.Wait()
	l.Release()

	a.True((<-ch).Delayed())
	a.Equal(int64(0), l.Count())
}

func TestWaitHold(t *testing.T) {
	a := assert.New(t)
	l := New()

	l.Hold()

	ch := l.WaitHold(2)
	l.Release()
	a.True((<-ch).Delayed())
	a.Equal(int64(2), l.Count())

	a.PanicsWithError("holds must be >= 0", func() {
		l.WaitHold(-1)
	})
}

func TestWaitLocked(t *testing.T) {
	a := assert.New(t)
	tl := &testLocker{}
	l := New(WithLocker(tl))

	if s := <-l.WaitLock(); a.False(s.Delayed()) && a.True(s.Locked()) {
		a.True(tl.Locked())
		// If the mutex weren't locked, there would be a fatal error.
		l.Unlock()
		a.False(tl.Locked())
	}

	l.Hold()

	ch := l.WaitLock()
	l.Release()
	if s := <-ch; a.True(s.Delayed()) && a.True(s.Locked()) {
		// If the mutex weren't locked, there would be a fatal error.
		l.Unlock()
	}
	l.Hold()
}

func TestOverRelease(t *testing.T) {
	a := assert.New(t)
	a.PanicsWithError("latch was over-released", func() { New().Release() })
}

type testLocker struct {
	mu     sync.Mutex
	locked int32
}

func (l *testLocker) Lock() {
	l.mu.Lock()
	atomic.StoreInt32(&l.locked, 1)
}

func (l *testLocker) Locked() bool {
	return atomic.LoadInt32(&l.locked) != 0
}

func (l *testLocker) Unlock() {
	l.mu.Unlock()
	atomic.StoreInt32(&l.locked, 0)
}
