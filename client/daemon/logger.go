package daemon

import (
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type ContextHook struct {
	ExcludeFile bool
	ExcludeFunc bool
	ExcludeLine bool
}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 3)
	n := runtime.Callers(6, pc)

	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, _ := frames.Next()
		if strings.Contains(frame.File, "github.com/sirupsen/logrus") {
			continue
		}

		// The entry.Data map must be copied before writing to, it is not
		// thread safe.
		data := make(map[string]interface{}, len(entry.Data)+3)
		for k, v := range entry.Data {
			data[k] = v
		}

		if !hook.ExcludeFile {
			data["file"] = path.Base(frame.File)
		}
		if !hook.ExcludeFunc {
			data["func"] = path.Base(frame.Function)
		}
		if !hook.ExcludeLine {
			data["line"] = frame.Line
		}

		entry.Data = data

		break
	}

	return nil
}

// NewLogger create logger instance
func NewLogger(logFilename string, debug bool) (*logrus.Logger, error) {
	log := logrus.New()
	log.Out = os.Stdout
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp:    true,
		QuoteEmptyFields: true,
	}
	log.Level = logrus.InfoLevel

	if debug {
		log.Level = logrus.DebugLevel
	}
	if logFilename != "" {
		hook, err := NewFileWriteHook(logFilename)
		if err != nil {
			return nil, err
		}

		log.Hooks.Add(hook)
	}

	log.Hooks.Add(ContextHook{
		ExcludeFunc: true,
	})

	return log, nil
}

// WriteHook is a logrus.Hook that logs to an io.Writer
type WriteHook struct {
	w         io.Writer
	formatter logrus.Formatter
}

// NewFileWriteHook returns a new WriteHook for a file
func NewFileWriteHook(filename string) (*WriteHook, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}

	return &WriteHook{
		w: f,
		formatter: &logrus.TextFormatter{
			FullTimestamp:    true,
			QuoteEmptyFields: true,
			DisableColors:    true,
		},
	}, nil
}

// NewStdoutWriteHook returns a new WriteHook for stdout
func NewStdoutWriteHook() *WriteHook {
	return &WriteHook{
		w: os.Stdout,
		formatter: &logrus.TextFormatter{
			FullTimestamp:    true,
			QuoteEmptyFields: true,
			DisableColors:    true,
		},
	}
}

// Levels returns Levels accepted by the WriteHook.
// All logrus.Levels are returned.
func (f *WriteHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire writes a logrus.Entry to the file
func (f *WriteHook) Fire(e *logrus.Entry) error {
	b, err := f.formatter.Format(e)
	if err != nil {
		return err
	}

	_, err = f.w.Write(b)
	return err
}
