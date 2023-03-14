# A Counting Latch

[![Golang](https://github.com/bobvawter/latch/actions/workflows/golang.yaml/badge.svg)](https://github.com/bobvawter/latch/actions/workflows/golang.yaml)
[![codecov](https://codecov.io/gh/bobvawter/latch/branch/main/graph/badge.svg?token=3lUp406eyx)](https://codecov.io/gh/bobvawter/latch)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bobvawter/latch)](https://pkg.go.dev/github.com/bobvawter/latch)

This package contains a notification-based, counter latch. This is
conceptually similar to a `sync.WaitGroup`, except that it does not
require foreknowledge of the total number of tasks that will be tracked.

The API is stable and semantically versioned.