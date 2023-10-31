<div align="center">
  <a href="https://github.com/OperantAI/secops-chaos/actions/workflows/build.yml">
    <img src="https://github.com/OperantAI/secops-chaos/actions/workflows/build.yml/badge.svg?branch=main">
  </a>
  <a href="https://github.com/operantai/secops-chaos/issues">
    <img src="https://img.shields.io/github/issues/operantai/secops-chaos">
  </a>
  <a href ="https://github.com/operantai/secops-chaos/issues">
    <img src="https://img.shields.io/github/issues/operantai/secops-chaos">
  </a>
</div>
<br />
<div align="center">
  <h3 align="center">secops-chaos</h3>
  <p align="center">
    Chaos-Driven Security Improvement
    <br />
    <a href="https://github.com/operantai/secops-chaos/blob/main/README.md"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md#reporting-bugs">Report Bug</a>
    ·
    <a href="https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md#suggesting-enhancements">Request Feature</a>
  </p>
</div>

**secops-chaos** is a Chaos Engineering tool focused on Security at Runtime. The tool was designed to discover security weaknesses in Cloud Native environments.

## Getting Started

### installation

``` sh
go install github.com/operantai/secops-chaos@latest
```

### Usage

``` sh
Usage:
  secops-chaos [command]

Available Commands:
  clean       Clean up after an experiment run
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Run an experiment
  verify      Verify the outcome of an experiment
  version     Output CLI version information

Flags:
  -h, --help   help for secops-chaos

Use "secops-chaos [command] --help" for more information about a command.
```

## Contributing

Please read the contribution guidelines, [here][contributing-url].

## License 

Distributed under the [Apache License 2.0][license-url].

[contributing-url]: https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md
[license-url]: https://github.com/operantai/secops-chaos/blob/main/LICENSE
