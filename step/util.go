package step

import (
	"sync"

	"github.com/bitrise-io/go-utils/v2/log"
)

// ChannelWriter can be used in place of an io.Writer where we want to redirect the output to a logger in real time
type ChannelWriter struct {
	ch     chan []byte
	mu     sync.Mutex
	logger log.Logger
}

func NewChannelWriter(chSize int, logger log.Logger) *ChannelWriter {
	return &ChannelWriter{
		ch:     make(chan []byte, chSize),
		logger: logger,
	}
}

func (cw *ChannelWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Send data to the channel
	select {
	case cw.ch <- append([]byte(nil), p...):
	default:
		// Channel is full, drop the data or handle accordingly
		cw.logger.Warnf("Channel is full, dropping data")
	}

	// Also log the data using the logger
	cw.logger.Infof(string(p))

	return len(p), nil
}
