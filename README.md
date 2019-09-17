[![Codacy Badge](https://api.codacy.com/project/badge/Grade/fcf7f843772443d49541eeef28b93397)](https://app.codacy.com/app/efureev/cache?utm_source=github.com&utm_medium=referral&utm_content=efureev/cache&utm_campaign=Badge_Grade_Dashboard)
[![Build Status](https://travis-ci.org/efureev/cache.svg?branch=master)](https://travis-ci.org/efureev/cache)
[![Maintainability](https://api.codeclimate.com/v1/badges/baa3f34d41f82f4510ce/maintainability)](https://codeclimate.com/github/efureev/cache/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/baa3f34d41f82f4510ce/test_coverage)](https://codeclimate.com/github/efureev/cache/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/efureev/cache)](https://goreportcard.com/report/github.com/efureev/cache)
[![codecov](https://codecov.io/gh/efureev/cache/branch/master/graph/badge.svg)](https://codecov.io/gh/efureev/cache)

# cache

cache is an in-memory key:value store/cache for applications running on a single machine. 
Its major advantage is that, being essentially a thread-safe `map[string]interface{}` with expiration
times, it doesn't need to serialize or transmit its contents over the network.

Any object can be stored, for a given duration or forever, and the cache can be
safely used by multiple goroutines.

### Installation

`go get -u github.com/efureev/cache/v1`

### Usage

```go
package main
import (
	"fmt"
	"github.com/efureev/cache"
	"time"
)

func main() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	c := cache.New(5*time.Minute, 10*time.Minute)

	// Set the value of the key "foo" to "bar", with the default expiration time
	c.Set("foo", "bar", cache.DefaultExpiration)

	// Set the value of the key "baz" to 42, with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("baz")
	c.Set("baz", 42, cache.NoExpiration)

	// Get the string associated with the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Println(foo)
	}

	// Since Go is statically typed, and cache values can be anything, type
	// assertion is needed when values are being passed to functions that don't
	// take arbitrary types, (i.e. interface{}). The simplest way to do this for
	// values which will only be used once--e.g. for passing to another
	// function--is:
	foo, found := c.Get("foo")
	if found {
		MyFunction(bar.(string))
	}

	// This gets tedious if the value is used several times in the same function.
	// You might do either of the following instead:
	if x, found := c.Get("foo"); found {
		foo := x.(string)
		// ...
	}
	// or
	var foo string
	if x, found := c.Get("foo"); found {
		foo = x.(string)
	}
	// ...
	// foo can then be passed around freely as a string

	// Want performance? Store pointers!
	c.Set("foo", &MyStruct, cache.DefaultExpiration)
	if x, found := c.Get("foo"); found {
		foo := x.(*MyStruct)
			// ...
	}
}
```
