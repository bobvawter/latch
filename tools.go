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

//go:build tools

//go:generate go run github.com/cockroachdb/crlfmt -w -ignore _gen.go .
//go:generate go run golang.org/x/lint/golint -set_exit_status ./...
//go:generate go run honnef.co/go/tools/cmd/staticcheck -checks all ./...

package main

import (
	_ "github.com/cockroachdb/crlfmt"
	_ "golang.org/x/lint/golint"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
