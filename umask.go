//go:build !windows

package main

import "syscall"

func umask(mask int) int {
	return syscall.Umask(mask)
}
