# MobileLab CI example

This example starts the Core without its dashboard, waits for the authenticated control plane to become ready, executes the same scenario directory with the fake device adapter, and retains JUnit XML, standalone HTML, the Core log, and a sample API response.

Build MobileLab at the repository root, then run:

```sh
make ci-example
```

Provider-ready definitions are included for GitHub Actions, GitLab CI, Azure Pipelines, and Jenkins. Copy the relevant file to the provider's conventional location. Each definition publishes the JUnit result and keeps the complete `artifacts/` directory. GitHub, GitLab, and Jenkins declare retention in their definitions; Azure artifact retention remains an honest pipeline/project setting because `PublishPipelineArtifact` has no per-task retention field.

The fake adapter validates Core/scenario integration without requiring a device farm. Replace `--platform fake` in `run.sh` and supply `--device`/`--app-id` only on runners where a booted emulator or simulator is deliberately provisioned.
