package step_test

import (
	"testing"

	"github.com/bitrise-steplib/bitrise-step-docker-build-push/step"
	"github.com/stretchr/testify/require"
)

func Test_ParseExtraOptions(t *testing.T) {
	cases := map[string]struct {
		given string
		want  []string
	}{
		"empty": {
			given: "",
			want:  nil,
		},
		"simple space seperated commands": {
			given: "--health-cmd pg_isready --build-arg foo=bar",
			want:  []string{"--health-cmd", "pg_isready", "--build-arg", "foo=bar"},
		},
		"commands using equal sign": {
			given: "--health-cmd=pg_isready --build-arg=foo=bar",
			want:  []string{"--health-cmd=pg_isready", "--build-arg=foo=bar"},
		},
		"commands using values with spaces": {
			given: "--health-cmd \"redis-cli ping\" --build-arg \"foo=bar and another\"",
			want:  []string{"--health-cmd", "redis-cli ping", "--build-arg", "foo=bar and another"},
		},
		"mixing commands with equal sign and without": {
			given: "--health-cmd=pg_isready --build-arg foo=bar --build-arg \"this=is something else\"",
			want:  []string{"--health-cmd=pg_isready", "--build-arg", "foo=bar", "--build-arg", "this=is something else"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := step.ParseExtraOptions(c.given)
			require.Equal(t, c.want, got)
		})
	}
}
