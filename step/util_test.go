package step

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestLogger struct {
	t      *testing.T
	output []string
}

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{
		output: []string{},
		t:      t,
	}
}

func (tl *TestLogger) Infof(format string, v ...interface{}) {
	tl.output = append(tl.output, format)
	tl.t.Logf(format, v...)
}

const multiLineUnicodeInput = `ᚠᛇᚻ᛫ᛒᛦᚦ᛫ᚠᚱᚩᚠᚢᚱ᛫ᚠᛁᚱᚪ᛫ᚷᛖᚻᚹᛦᛚᚳᚢᛗ
ᛋᚳᛖᚪᛚ᛫ᚦᛖᚪᚻ᛫ᛗᚪᚾᚾᚪ᛫ᚷᛖᚻᚹᛦᛚᚳ᛫ᛗᛁᚳᛚᚢᚾ᛫ᚻᛦᛏ᛫ᛞᚫᛚᚪᚾ
ᚷᛁᚠ᛫ᚻᛖ᛫ᚹᛁᛚᛖ᛫ᚠᚩᚱ᛫ᛞᚱᛁᚻᛏᚾᛖ᛫ᛞᚩᛗᛖᛋ᛫ᚻᛚᛇᛏᚪᚾ᛬`

func Test_LogWriter(t *testing.T) {
	cases := map[string]struct {
		input                       string
		expectedNumberOfLinesLogged int
	}{
		"empty": {
			input:                       "",
			expectedNumberOfLinesLogged: 0,
		},
		"only/\n": {
			input:                       "\n\n",
			expectedNumberOfLinesLogged: 2,
		},
		"only/\r": {
			input:                       "\r\r",
			expectedNumberOfLinesLogged: 2,
		},
		"only/\r/\n": {
			input:                       "\r\n\r\n",
			expectedNumberOfLinesLogged: 4,
		},
		"onelineinput": {
			input:                       "onelineinput",
			expectedNumberOfLinesLogged: 1,
		},
		"multilineinput": {
			input:                       "line1\nline2\rline3\nline4\n",
			expectedNumberOfLinesLogged: 4,
		},
		multiLineUnicodeInput: {
			input:                       multiLineUnicodeInput,
			expectedNumberOfLinesLogged: 3,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			tl := TestLogger{t: t}
			lw := NewLoggerWriter(&tl)

			_, err := lw.Write([]byte(c.input))
			require.NoError(t, err)
			lw.Flush()

			if len(tl.output) != c.expectedNumberOfLinesLogged {
				t.Errorf("expected %d lines logged, got %d", c.expectedNumberOfLinesLogged, len(tl.output))
			}
		})
	}
}
