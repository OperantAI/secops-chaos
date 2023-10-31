<div align="center">
  [![Build][build-shield]][build-url]
  [![Issues][issues-shield]][issues-url]
  [![License][license-shield]][license-url]
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
    <a href="https://github.com/operantai/secops-chaos/issues">Report Bug</a>
    ·
    <a href="https://github.com/operantai/secops-chaos/issues">Request Feature</a>
  </p>
</div>

`secops-chaos` is a Chaos Engineering tool focused on Security at Runtime. The tool was designed to discover security weaknesses in Cloud Native environments.

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

[build-shield]: https://github.com/OperantAI/secops-chaos/actions/workflows/build.yml/badge.svg?branch=main
[build-url]: https://github.com/OperantAI/secops-chaos/actions/workflows/build.yml
[issues-shield]: https://img.shields.io/github/issues/operantai/secops-chaos
[issues-url]: https://github.com/operantai/secops-chaos/issues
[license-shield]: https://img.shields.io/github/license/operantai/secops-chaos
[license-url]: https://github.com/operantai/secops-chaos/blob/master/LICENSE
[contributing-url]: https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md
