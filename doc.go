// Package goblip provides a lightweight, dependency-free file-watching
// utility for Go development. It automatically detects source changes
// and restarts a running process, similar to nodemon or air.
//
// GoBlip is designed for fast, simple local development of Go applications,
// web servers, or CLI tools. It works entirely with the Go standard library
// and uses polling (not fsnotify) for maximum cross-platform compatibility.
//
// Typical usage:
//
//   goblip -- go run .
//   goblip -v -- go run main.go
//
// It is commonly used with web frameworks such as Gin or you can use it with any go project:
//
//   goblip -- go run .
//   # automatically restarts your Gin web server when .go or .html files change
//
// Default watched extensions include .go, .mod, .sum, .html, .tpl, .css, and .js,
// but can be customized using the -ext flag. The polling interval can be
// controlled with the -interval flag.
//
// Example:
//
//   goblip -ext ".go,.tpl" -interval 300ms -- go run ./cmd/server
//
// GoBlip is framework-agnostic — it works equally well with Gin, Fiber, Echo,
// Chi, or any other Go application that you want to auto-restart on file changes.
//
// Author: Ayush Srivastava
// License: MIT
package goblip
