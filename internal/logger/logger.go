package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	InfoLogger      *log.Logger
	DebugLogger     *log.Logger
	ErrorLogger     *log.Logger
	SeparatorLogger *log.Logger
	HeadingLogger   *log.Logger
)

func InitLogger(logfile string, logLevel string) {
	var multi_writer io.Writer
	if logfile == "" {
		multi_writer = io.MultiWriter(os.Stdout)
	} else {
		file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatal(err)
		}
		multi_writer = io.MultiWriter(os.Stdout, file)
	}

	var debug_writer io.Writer
	var info_writer io.Writer

	if logLevel == "debug" {
		debug_writer = multi_writer
		info_writer = multi_writer
	} else if logLevel == "info" {
		debug_writer = io.Discard
		info_writer = multi_writer
	} else {
		debug_writer = io.Discard
		info_writer = io.Discard
	}

	InfoLogger = log.New(info_writer, "INFO: ", 0)
	ErrorLogger = log.New(multi_writer, "ERROR: ", 0)
	DebugLogger = log.New(debug_writer, "DEBUG: ", 0)
	SeparatorLogger = log.New(multi_writer, "", 0)
	HeadingLogger = log.New(multi_writer, "", log.Ldate|log.Ltime)
}

func format_string(format string, args ...interface{}) string {
	args2 := make([]string, len(args))
	for i, v := range args {
		if i%2 == 0 {
			args2[i] = fmt.Sprintf("{%v}", v)
		} else {
			args2[i] = fmt.Sprint(v)
		}
	}
	r := strings.NewReplacer(args2...)
	return r.Replace(format)
}

func write_to_log(loggerType int, format string, args ...interface{}) {

	log_msg := format_string(format, args...)
	switch loggerType {
	case 1:
		ErrorLogger.Println(log_msg)
	case 2:
		InfoLogger.Println(log_msg)
	case 3:
		// below adds caller info to the string to be logged
		// _, fn, line, _ := runtime.Caller(1)
		// format = filepath.Base(fn) + ":" + strconv.Itoa(line) + ": " + format
		DebugLogger.Println(log_msg)
	case 4:
		SeparatorLogger.Println(log_msg)
	case 5:
		HeadingLogger.Println(log_msg)
	}
}

func Error(format string, args ...interface{}) {
	write_to_log(1, format, args...)
}

func Info(format string, args ...interface{}) {
	write_to_log(2, format, args...)
}

func Debug(format string, args ...interface{}) {
	write_to_log(3, format, args...)
}

func Separate(format string, args ...interface{}) {
	write_to_log(4, format, args...)
}

func Heading(format string, args ...interface{}) {
	write_to_log(5, format, args...)
}
