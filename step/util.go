package step

type InfoLogger interface {
	Infof(format string, v ...interface{})
}

// LogWriter can be used in place of an io.Writer where we want to redirect the output to a logger in real time
type LogWriter struct {
	buffer []byte
	logger InfoLogger
}

func NewLoggerWriter(logger InfoLogger) *LogWriter {
	return &LogWriter{
		logger: logger,
	}
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	for index := range p {
		if p[index] == '\n' || p[index] == '\r' {
			lw.logger.Infof(string(lw.buffer))
			lw.buffer = nil
		} else {
			lw.buffer = append(lw.buffer, p[index])
		}
	}
	return len(p), nil
}

func (lw *LogWriter) Flush() {
	if len(lw.buffer) > 0 {
		lw.logger.Infof(string(lw.buffer))
		lw.buffer = nil
	}
}
