//go:build !windows

package main

import "syscall"

func umask() {
	syscall.Umask(0)
}
