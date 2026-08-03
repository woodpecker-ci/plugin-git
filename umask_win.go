//go:build windows

package main

func umask(mask int) int {
	return 0
}
