package step

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/cache"
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

const (
	dockerCacheKeyTemplate = "docker-%s-{{ .OS }}-{{ .Arch }}-{{ .Branch }}-{{ .CommitHash }}"
	dockerCacheFolder      = "/tmp/.buildx-cache"
)

func New(
	logger log.Logger,
	inputParser stepconf.InputParser,
	commandFactory command.Factory,
	pathChecker pathutil.PathChecker,
	pathProvider pathutil.PathProvider,
	pathModifier pathutil.PathModifier,
	envRepo env.Repository,
) DockerBuildPushStep {
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

	tags := strings.Split(input.Tags, "\n")
	tagUsedInCacheKey := tags[0]

	if input.UseBitriseCache {
		if err := step.restoreCache(input, tagUsedInCacheKey); err != nil {
			return fmt.Errorf("failed to restore cache: %w", err)
		}
	}

	if err := step.dockerBuild(input); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	if input.UseBitriseCache {
		if err := step.saveCache(input, tagUsedInCacheKey); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}
	}
	return nil
}

func (step DockerBuildPushStep) restoreCache(input Input, imageName string) error {
	step.logger.Infof("Restoring cache...")
	saver := cache.NewRestorer(step.envRepo, step.logger, step.commandFactory)

	var cacheKey = []string{
		fmt.Sprintf(dockerCacheKeyTemplate, imageName),
		fmt.Sprintf("docker-%s-{{ .OS }}-{{ .Arch }}-{{ .Branch }}", imageName),
		fmt.Sprintf("docker-%s-{{ .OS }}-{{ .Arch }}", imageName),
	}

	return saver.Restore(cache.RestoreCacheInput{
		StepId:  "restore-cache",
		Verbose: input.Verbose,
		Keys:    cacheKey,
	})
}

func (step DockerBuildPushStep) dockerBuild(input Input) error {
	step.logger.Infof("Building docker image...")
	//cmd := step.commandFactory.Create("docker", []string{"build", "TODO"}, nil)
	//out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	//if err != nil {
	//	return fmt.Errorf("failed to build docker image %s: %w", out, err)
	//}

	cmd := step.commandFactory.Create("mkdir", []string{"-p", dockerCacheFolder}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create cache folder %s: %w", out, err)
	}

	cmd = step.commandFactory.Create("touch", []string{fmt.Sprintf("%s/kutyacica.fake", dockerCacheFolder)}, nil)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create cache folder %s: %w", out, err)
	}

	return nil
}

func (step DockerBuildPushStep) saveCache(input Input, imageName string) error {
	step.logger.Infof("Saving cache...")
	saver := cache.NewSaver(step.envRepo, step.logger, step.pathProvider, step.pathModifier, step.pathChecker)

	return saver.Save(cache.SaveCacheInput{
		StepId:      "save-cache",
		Verbose:     input.Verbose,
		Key:         fmt.Sprintf(dockerCacheKeyTemplate, imageName),
		Paths:       []string{dockerCacheFolder},
		IsKeyUnique: false,
	})
}
