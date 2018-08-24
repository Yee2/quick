package main

import (
	"fmt"
	"runtime"
	"strings"
	"strconv"
	"time"
	"path/filepath"
)

type Color string

const (
	ANSI_RED    Color = "[01;31m"
	ANSI_WHITE  Color = "[01;37m"
	ANSI_GREEN  Color = "[01;32m"
	ANSI_YELLOW Color = "[01;33m"
	ANSI_BLUE   Color = "[01;34m"
	ANSI_BELL   Color = "0"
	ANSI_BLINK  Color = "[05m"
	ANSI_REV    Color = "[07m"
	ANSI_UL     Color = "[04m"
)

func Show(c Color, format string, a ...interface{}) {
	fmt.Printf("\x1B%s%s\x1B[0m", c, fmt.Sprintf(format, a...))
}


func logf(format string, a ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Show(ANSI_GREEN, "- %-34s%s -\n\n",
		filepath.Base(file)+":"+strconv.Itoa(line),
		time.Now().Format("15:04:05"),
	)
	Show(ANSI_YELLOW, format, a...)
	Show(ANSI_GREEN, "\n-%s-\n", strings.Repeat("-", 44))
}