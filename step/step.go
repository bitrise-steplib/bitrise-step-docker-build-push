package step

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

type Input struct {
	Tags            string `env:"tags,required"`
	UseBitriseCache bool   `env:"use_bitrise_cache"`
	Push            bool   `env:"push"`
	Verbose         bool   `env:"verbose,required"`
}

type DockerBuildPushStep struct {
	logger         log.Logger
	inputParser    stepconf.InputParser
	commandFactory command.Factory
	pathChecker    pathutil.PathChecker
	pathProvider   pathutil.PathProvider
	pathModifier   pathutil.PathModifier
	envRepo        env.Repository
}

func New(logger log.Logger, inputParser stepconf.InputParser, commandFactory command.Factory, pathChecker pathutil.PathChecker, pathProvider pathutil.PathProvider, pathModifier pathutil.PathModifier, envRepo env.Repository) DockerBuildPushStep {
	return DockerBuildPushStep{
		logger:         logger,
		inputParser:    inputParser,
		commandFactory: commandFactory,
		pathChecker:    pathChecker,
		pathProvider:   pathProvider,
		pathModifier:   pathModifier,
		envRepo:        envRepo,
	}
}

func (step DockerBuildPushStep) Run() error {
	var input Input
	if err := step.inputParser.Parse(&input); err != nil {
		return fmt.Errorf("failed to parse inputs: %w", err)
	}
	stepconf.Print(input)
	step.logger.Println()

	step.logger.EnableDebugLog(input.Verbose)

	return nil
	// saver := cache.NewSaver(step.envRepo, step.logger, step.pathProvider, step.pathModifier, step.pathChecker)

	//return saver.Save(cache.SaveCacheInput{
	//	StepId:      "save-cache",
	//	Verbose:     input.Verbose,
	//	Key:         "TODO",
	//	Paths:       strings.Split("TODO", "\n"),
	//	IsKeyUnique: false,
	//})
}
