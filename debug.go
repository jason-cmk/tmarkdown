package main

import (
	"fmt"
	"os"
	"time"
)

// Set TMARKDOWN_DEBUG to a file path to log input events there.
var debugFile = func() *os.File {
	p := os.Getenv("TMARKDOWN_DEBUG")
	if p == "" {
		return nil
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil
	}
	return f
}()

func debugf(format string, args ...any) {
	if debugFile == nil {
		return
	}
	fmt.Fprintf(debugFile, time.Now().Format("15:04:05.000")+" "+format+"\n", args...)
}
