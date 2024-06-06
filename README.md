<div align="center">
  <a href="https://github.com/OperantAI/secops-chaos/actions/workflows/build.yml">
    <img src="https://github.com/OperantAI/secops-chaos/actions/workflows/build.yml/badge.svg?branch=main">
  </a>
  <a href="https://github.com/operantai/secops-chaos/issues">
    <img src="https://img.shields.io/github/issues/operantai/secops-chaos">
  </a>
  <a href ="https://github.com/operantai/secops-chaos/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/operantai/secops-chaos">
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

### Installation

You can fetch the latest release [here][latest-release-url], or you can build from source.

#### Building from Source

To build from source, you'll need to have [Go](https://golang.org/) installed.

```sh
git clone https://github.com/operantai/secops-chaos
cd secops-chaos
make build
```

### Usage

The design of **secops-chaos** can be broken down into three concepts:

- **Experiments** - Experiments actively try to run something to discover if a security weakness is present.
- **Verifiers** - Verifiers look at the results of an Experiment and reports their outcome.
- **Components** - Components are additional applications installed on a K8s cluster to enable and enhance experiment functionality.


The secops-chaos CLI mirrors this, and exposes `run`, `verify` and `component` commands.

#### Experiments & Verifiers

To start, you need to run an experiment.

Each experiment is defined by a `experiment` file which allows you to tweak your experiment parameters to suit your scenarios.

For a full list of experiments you can run, see the [experiments][experiments-dir-url] directory.

``` sh
secops-chaos experiment run -f experiments/host_path_volume.yaml
```

Once you've successfully run the experiment, you can verify if it was sucessful or not:

```sh
secops-chaos experiment verify -f experiments/host_path_volume.yaml
```

You can also output a JSON with the verifier results by using the `-j` flag.

#### Components

Some experiments require additional applications installed to run or enhance their functionality.

These can be added by providing a YAML file, see the [components](components-dir-url) directory for examples.

```sh
secops-chaos component install -f components/secops-chaos-ai.yaml
```

```sh
secops-chaos component uninstall -f components/secops-chaos-ai.yaml
```

Experiments that need a component will warn you, and allow for installation during runtime.

## Contributing

Please read the contribution guidelines, [here][contributing-url].

## License 

Distributed under the [Apache License 2.0][license-url].

[latest-release-url]: https://github.com/operantai/secops-chaos/releases/latest
[experiments-dir-url]: https://github.com/operantai/secops-chaos/blob/main/experiments
[components-dir-url]: https://github.com/operantai/secops-chaos/blob/main/components
[contributing-url]: https://github.com/operantai/secops-chaos/blob/main/CONTRIBUTING.md
[license-url]: https://github.com/operantai/secops-chaos/blob/main/LICENSE
