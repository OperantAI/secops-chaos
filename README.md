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
    Security-focused Chaos Experiments for DevSecOps Teams
    <br />
    <a href="https://github.com/operantai/secops-chaos/blob/main/README.md"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md#reporting-bugs">Report Bug</a>
    ·
    <a href="https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md#suggesting-enhancements">Request Feature</a>
  </p>
</div>

**secops-chaos** is a Chaos Engineering tool focused on Security at Runtime. The tool is designed to discover security weaknesses by experimentation in Cloud Native environments.

## Getting Started

### installation

``` sh
go install github.com/operantai/secops-chaos@latest
```

Go will automatically install it in your `$GOPATH/bin` directory, which should be in your `$PATH`.

### Usage

The design of **secops-chaos** can be broken down into two components:

- **Experiments** - Experiments actively try to run something to discover if a security weakness is present.
- **Verifiers** - Verifiers look at the results of an Experiment and reports their outcome.

The secops-chaos CLI mirrors this, and exposes `run` & `verify` commands. To start, you need to run an experiment:

``` sh
secops-chaos run -f experiments/host_path_volume.yaml
```

Once you've successfully run the experiment, you can verify if it was sucessful or not:

```sh
secops-chaos verify -f experiments/host_path_volume.yaml
```

You can also output a JSON with the verifier results by using the `-j` flag.

For a full list of experiments you can run, see the [experiments][experiments-dir-url] directory.

## Contributing

Please read the contribution guidelines, [here][contributing-url].

## License 

Distributed under the [Apache License 2.0][license-url].

[experiments-dir-url]: https://github.com/operantai/secops-chaos/blob/main/experiments
[contributing-url]: https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md
[license-url]: https://github.com/operantai/secops-chaos/blob/main/LICENSE
