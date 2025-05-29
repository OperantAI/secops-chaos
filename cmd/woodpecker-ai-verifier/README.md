# Woodpecker for AI

Woodpecker component for redteaming against AI apps and APIs

## Pre-commit

```sh
pip3 install pre-commit
pre-commit run --files ./*
````

## Build

```sh
pip3 install .
```

## Running

```sh
./entrypoint.sh
````

## Docker Build

```sh
docker build -t woodpecker-ai-verifier:latest . -f ./build/Dockerfile.woodpecker-ai-verifier
````

## Docker Run

```sh
docker run -p 8000:8000 woodpecker-ai-verifier:latest
````
