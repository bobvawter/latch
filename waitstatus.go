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

// WaitStatus is returned from the Wait functions.
type WaitStatus int

// Delayed returns true if the call to Wait was delayed by existing
// holds.
func (s WaitStatus) Delayed() bool {
	return s&delayed == delayed
}

// Locked returns true if the call to wait left the Counter in a locked
// state.
func (s WaitStatus) Locked() bool {
	return s&locked == locked
}

func (s WaitStatus) locked() WaitStatus {
	return s | locked
}

const (
	locked WaitStatus = 1 << iota
	delayed
)
