package glockmon

import (
	_ "bytes"
	"github.com/cespare/xxhash/v2"
	"runtime"
)

func getStackTrace() string {
	for size := 1024; size <= 65536; size *= 2 {
		buf := make([]byte, size)
		n := runtime.Stack(buf, false)
		if n < size {
			return string(buf[:n])
		}
	}
	return "stack trace too deep"
}

func hashStack(stack string) uint64 {
	return xxhash.Sum64String(stack)
}
