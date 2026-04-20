# Docker Build & Push

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/bitrise-step-docker-build-push?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/bitrise-step-docker-build-push/releases)

Building and pushing docker images with built-in cache support


<details>
<summary>Description</summary>

Enables you to build and optionally push docker images with built-in Bitrise key-value cache support.
It is possible to customize the docker build command to utilize other cache mechanisms, such as registry-cache.

</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `tags` | List of tags (full image names) to be applied to the built image  Add one tag per line. Example: `myregistry.com/myimage:latest` | required |  |
| `context` | Path to the build context to be used for the docker build  The path is relative to the working directory. | required | `.` |
| `file` | Path to the Dockerfile to be built  The path is relative to the working directory. | required | `./Dockerfile` |
| `push` | When set to 'true', the docker image will be pushed. | required | `false` |
| `use_bitrise_cache` | When set to 'true', the docker image will automatically be cached using Bitrise key-value cache.  The following cache keys will be used: - docker-imagename-{{ .OS }}-{{ .Arch }}-{{ .Branch }}-{{ .CommitHash }} - docker-imagename-{{ .OS }}-{{ .Arch }}-{{ .Branch }} - docker-imagename-{{ .OS }}-{{ .Arch }}  Warning: Do not specify the cache-to and cache-from parameters when using this option. | required | `false` |
| `build_arg` | List of build arguments to be passed to the docker build  Add one build argument per line. Example: `MY_BUILD_ARG=myvalue` |  |  |
| `cache_from` | List of arguments to be passed to the docker build for cache-from  Add one cache-from argument per line. Example: `type=registry,ref=myregistry.dev/myrepository/my-image:latest` |  |  |
| `cache_to` | List of arguments to be passed to the docker build for cache-to  Add one cache-to argument per line. Example: `type=registry,ref=myregistry.dev/myrepository/my-image:latest,mode=max,compression=zstd` |  |  |
| `extra_options` | List of extra options to be passed to the docker build  Extra options must be in the format of `--option value` or `--option=value`.  Warning:   When using values with quotes in them (for example when they contain spaces) do not use the equal sign.   Separate it with spaces instead. Example: `--option "value with spaces"`  Add one extra option per line. |  |  |
| `buildx_host_network` | Enables to use the host network with the buildkit build container | required | `false` |
| `verbose` | Enable logging additional information for troubleshooting | required | `false` |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/bitrise-step-docker-build-push/pulls) and [issues](https://github.com/bitrise-steplib/bitrise-step-docker-build-push/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
