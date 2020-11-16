# A Counting Latch

[![Build Status](https://travis-ci.com/bobvawter/latch.svg?branch=main)](https://travis-ci.com/bobvawter/latch)
[![codecov](https://codecov.io/gh/bobvawter/latch/branch/main/graph/badge.svg)](https://codecov.io/gh/bobvawter/latch)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bobvawter/latch)](https://pkg.go.dev/github.com/bobvawter/latch)

This package contains a notification-based, counter latch. This is
conceptually similar to a `sync.WaitGroup`, except that it does not
require foreknowledge of the total number of tasks that will be tracked.

The API is stable and semantically versioned.