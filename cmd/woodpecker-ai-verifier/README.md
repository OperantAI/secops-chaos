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
export OPENAI_API_KEY="my-api-key"
./entrypoint.sh
````

## Docker Build

```sh
docker build -t woodpecker-ai:latest . -f Dockerfile
````

## Docker Run

```sh
docker run -p 8000:8000 -e OPENAI_KEY=<> woodpecker-ai:latest
````
