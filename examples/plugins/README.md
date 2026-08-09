# MobileLab plugin example

This project-local `echo` plugin demonstrates the versioned `mobilelab.plugin/v1` protocol and the public Go authoring package at `pkg/plugin`. It does not need a running MobileLab sandbox.

From the repository root:

```sh
make build plugin-example
cd examples/plugins
../../bin/mobilelab plugin list
../../bin/mobilelab plugin inspect echo
../../bin/mobilelab plugin run echo echo --input input.json
```

MobileLab discovers only `mobilelab/plugins/<name>/plugin.yaml` beneath the current project. It never downloads or runs a plugin during discovery, `start`, or scenario execution. `plugin run` is the explicit execution boundary.

The executable receives one strict JSON request on standard input and must return one strict JSON response on standard output. Diagnostic output belongs on standard error. Keep standard output reserved for the protocol.

Plugins are trusted local executables. MobileLab limits execution time, request/response size, and inherited environment variables, but v0.7 does not provide an operating-system sandbox. Review the executable and compare the SHA-256 fingerprint shown by `plugin inspect` before running third-party code.
