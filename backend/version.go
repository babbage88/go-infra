package main

import (
	"fmt"
	"runtime"
)

const (
	Author  = "Justin Trahan"
	Version = "v1.0.4"
	Name    = "go-infra"
)

func showVersion() string {
	fmt.Printf("%s\nVersion: %s\nOS: %s Arch: %s\nAuthor: %s\n", Name, Version, runtime.GOOS, runtime.GOARCH, Author)

	return Version
}
