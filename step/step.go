package step

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/cache"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

type Input struct {
	UseBitriseCache   bool `env:"use_bitrise_cache,required"`
	Push              bool `env:"push,required"`
	Verbose           bool `env:"verbose,required"`
	BuildxHostNetwork bool `env:"buildx_host_network,required"`

	Tags         string `env:"tags,required"`
	File         string `env:"file,required"`
	Context      string `env:"context,required"`
	BuildArg     string `env:"build_arg"`
	CacheFrom    string `env:"cache_from"`
	CacheTo      string `env:"cache_to"`
	ExtraOptions string `env:"extra_options"`
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
	dockerCacheKeyTemplate     = "docker-%s-{{ .OS }}-{{ .Arch }}-{{ .Branch }}-{{ .CommitHash }}"
	dockerCacheFolder          = "/tmp/.buildx-cache"
	dockerCacheFolderTemporary = "/tmp/.buildx-cache-new"
	stepId                     = "docker-build-push"
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
		return fmt.Errorf("parse inputs: %w", err)
	}
	stepconf.Print(input)
	step.logger.Println()

	step.logger.EnableDebugLog(input.Verbose)

	tagUsedInCacheKey := strings.Split(input.Tags, "\n")[0]

	if input.UseBitriseCache {
		if err := step.restoreCache(input, tagUsedInCacheKey); err != nil {
			return fmt.Errorf("restore cache: %w", err)
		}
	}

	if err := step.dockerBuild(input); err != nil {
		return fmt.Errorf("build docker image: %w", err)
	}

	if input.UseBitriseCache {
		if err := step.saveCache(input, tagUsedInCacheKey); err != nil {
			return fmt.Errorf("save cache: %w", err)
		}
	}
	return nil
}

func (step DockerBuildPushStep) restoreCache(input Input, imageName string) error {
	step.logger.Infof("Restoring cache...")
	restorer := cache.NewRestorer(step.envRepo, step.logger, step.commandFactory)

	var cacheKey = []string{
		fmt.Sprintf(dockerCacheKeyTemplate, imageName),
		fmt.Sprintf("docker-%s-{{ .OS }}-{{ .Arch }}-{{ .Branch }}", imageName),
		fmt.Sprintf("docker-%s-{{ .OS }}-{{ .Arch }}", imageName),
	}

	return restorer.Restore(cache.RestoreCacheInput{
		StepId:  stepId,
		Verbose: input.Verbose,
		Keys:    cacheKey,
	})
}

func (step DockerBuildPushStep) saveCache(input Input, imageName string) error {
	step.logger.Infof("Saving cache...")
	saver := cache.NewSaver(step.envRepo, step.logger, step.pathProvider, step.pathModifier, step.pathChecker)

	return saver.Save(cache.SaveCacheInput{
		StepId:      stepId,
		Verbose:     input.Verbose,
		Key:         fmt.Sprintf(dockerCacheKeyTemplate, imageName),
		Paths:       []string{dockerCacheFolder},
		IsKeyUnique: false,
	})
}

func (step DockerBuildPushStep) dockerBuild(input Input) error {
	step.logger.Infof("Building docker image...")

	if err := step.createCacheFolder(dockerCacheFolder); err != nil {
		return fmt.Errorf("create cache folder: %w", err)
	}
	if err := step.createCacheFolder(dockerCacheFolderTemporary); err != nil {
		return fmt.Errorf("create cache folder: %w", err)
	}

	buildkitContainer, err := step.initializeBuildkit(input)
	if err != nil {
		return fmt.Errorf("initialize buildkit: %w", err)
	}
	defer func() {
		if err := step.destroyContainer(buildkitContainer); err != nil {
			step.logger.Errorf("destroy buildx instance: %s", err)
		}
	}()

	if err := step.build(input); err != nil {
		return fmt.Errorf("build docker image: %w", err)
	}

	// To make sure the cache is not growing indefinitely,
	// we remove the original cache folder and move the new one to its place
	if err := step.moveCacheFolder(dockerCacheFolderTemporary, dockerCacheFolder); err != nil {
		return fmt.Errorf("move cache folder: %w", err)
	}

	return nil
}

func (step DockerBuildPushStep) destroyContainer(container string) error {
	args := []string{
		"buildx", "rm", "--force", container,
	}
	cmd := step.commandFactory.Create("docker", args, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("remove buildx instance %s: %w", out, err)
	}
	return nil
}

func (step DockerBuildPushStep) build(input Input) error {
	args := []string{
		"buildx",
		"build",
	}

	if input.BuildArg != "" {
		for _, arg := range strings.Split(input.BuildArg, "\n") {
			args = append(args, "--build-arg", arg)
		}
	}

	switch {
	case input.UseBitriseCache:
		args = append(args, fmt.Sprintf("--cache-from=type=local,src=%s", dockerCacheFolder))
		args = append(args, fmt.Sprintf("--cache-to=type=local,dest=%s,mode=max,compression=zstd", dockerCacheFolderTemporary))
	case input.CacheFrom != "":
		for _, cacheFrom := range strings.Split(input.CacheFrom, "\n") {
			args = append(args, fmt.Sprintf("--cache-from=%s", cacheFrom))
		}
		fallthrough
	case input.CacheTo != "":
		for _, cacheTo := range strings.Split(input.CacheTo, "\n") {
			args = append(args, fmt.Sprintf("--cache-to=%s", cacheTo))
		}
	}

	if input.ExtraOptions != "" {
		for _, option := range strings.Split(input.ExtraOptions, "\n") {
			// This regex splits the string by spaces, but keeps quoted strings together
			// Example --build-arg "-X main.version=1.0.0" will be split into --build-arg and "-X main.version=1.0.0"
			r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
			result := r.FindAllString(option, -1)

			// Remove quotes from the strings
			var options []string
			for _, result := range result {
				options = append(options, strings.ReplaceAll(result, "\"", ""))
			}

			args = append(args, options...)
		}
	}

	if input.Push {
		args = append(args, "--push")
	} else {
		// The --load parameter is used to load the image into the local docker daemon
		// This is needed because the docker buildx build command will keep the result in cache only,
		// preventing the use of the image in the same build
		args = append(args, "--load")
	}

	for _, tag := range strings.Split(input.Tags, "\n") {
		args = append(args, "--tag", tag)
	}

	args = append(args, []string{"-f", input.File, input.Context}...)

	step.logger.Infof("$ docker %s", strings.Join(args, " "))

	buildxCmd := step.commandFactory.Create("docker", args, &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	})

	err := buildxCmd.Run()
	if err != nil {
		return fmt.Errorf("build docker image with buildx: %w", err)
	}

	return nil
}

func (step DockerBuildPushStep) initializeBuildkit(input Input) (string, error) {
	args := []string{
		"buildx", "create", "--use",
	}

	if input.BuildxHostNetwork {
		args = append(args, "--driver-opt", "network=host", "--buildkitd-flags", "--allow-insecure-entitlement network.host")
	}

	createCmd := step.commandFactory.Create("docker", args, nil)

	step.logger.Infof("$ docker %s", strings.Join(args, " "))

	out, err := createCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("create buildx instance %s: %w", out, err)
	}
	return out, nil
}

func (step DockerBuildPushStep) createCacheFolder(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("create cache folder %w", err)
	}

	return nil
}

func (step DockerBuildPushStep) moveCacheFolder(from string, to string) error {
	err := os.RemoveAll(to)
	if err != nil {
		return fmt.Errorf("remove cache folder %w", err)
	}

	err = os.Rename(from, to)
	if err != nil {
		return fmt.Errorf("move cache folder: %w", err)
	}

	return nil
}
