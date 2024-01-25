package step

import (
	"github.com/bitrise-io/go-utils/v2/log"
)

// LogWriter can be used in place of an io.Writer where we want to redirect the output to a logger in real time
type LogWriter struct {
	buffer []byte
	logger log.Logger
}

func NewLoggerWriter(logger log.Logger) *LogWriter {
	return &LogWriter{
		logger: logger,
	}
}

func (cw *LogWriter) Write(p []byte) (n int, err error) {
	for index := range p {
		if (p[index] == '\n' || p[index] == '\r') && len(cw.buffer) > 0 {
			cw.logger.Infof(string(cw.buffer))
			cw.buffer = nil
		} else {
			cw.buffer = append(cw.buffer, p[index])
		}
	}
	return len(p), nil
}

func (cw *LogWriter) Flush() {
	if len(cw.buffer) > 0 {
		cw.logger.Infof(string(cw.buffer))
		cw.buffer = nil
	}
}
