package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

func logf(format string, a ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("\x1B[01;33m%s %s[%d]:\x1B[0m",
		time.Now().Format("15:04:05"),
		filepath.Base(file),
		line,
	)
	fmt.Printf(format, a...)
	fmt.Println()
}
