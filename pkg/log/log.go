package log

import (
	"fmt"
	"log"
	"runtime"
)

func trace(n int) {
	i := 0
	for {
		pt, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		funcName := runtime.FuncForPC(pt).Name()
		fmt.Printf("file=%s, line=%d, func=%v\n", file, line, funcName)
		i++

		if i > n {
			break
		}
	}
	fmt.Println()
}

func Errorf(format string, v ...interface{}) {
	fmt.Println()
	log.Printf(format, v...)
	trace(100)
}

func Error(v ...interface{}) {
	fmt.Println()
	log.Println(v...)
	trace(100)
}

func Infof(format string, v ...interface{}) {
	fmt.Println()
	log.Printf(format, v...)
	trace(2)
}

func Info(v ...interface{}) {
	fmt.Println()
	log.Println(v...)
	trace(2)
}

func Debugf(format string, v ...interface{}) {
	fmt.Println()
	log.Printf(format, v...)
}

func Debugln(v ...interface{}) {
	fmt.Println()
	log.Println(v...)
}
