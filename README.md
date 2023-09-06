# dockerfiler

[![Build Status](https://github.com/craftslab/dockerfiler/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/craftslab/dockerfiler/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/craftslab/dockerfiler/branch/main/graph/badge.svg?token=7PMQALLZLY)](https://codecov.io/gh/craftslab/dockerfiler)
[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/dockerfiler)](https://goreportcard.com/report/github.com/craftslab/dockerfiler)
[![License](https://img.shields.io/github/license/craftslab/dockerfiler.svg)](https://github.com/craftslab/dockerfiler/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/dockerfiler.svg)](https://github.com/craftslab/dockerfiler/tags)



## Introduction

*dockerfiler* is a Dockerfile generator written in Go.



## Prerequisites

- Go >= 1.18.0



## Install

```bash
curl -LO https://storage.googleapis.com/container-diff/latest/container-diff-linux-amd64
sudo install container-diff-linux-amd64 /usr/local/bin/container-diff
```



## Run

```bash
container-diff diff --type=apt --type=node --type=pip --json daemon://image1 daemon://image2 > diff.json

version=latest make build
./bin/dockerfiler --input-file=diff.json --output-file=Dockerfile
```



## Usage

```
usage: dockerfiler --input-file=INPUT-FILE --output-file=OUTPUT-FILE [<flags>]

Dockerfile generator


Flags:
  --[no-]help                Show context-sensitive help (also try --help-long
                             and --help-man).
  --[no-]version             Show application version.
  --input-file=INPUT-FILE    Input file (.json)
  --output-file=OUTPUT-FILE  Output file (Dockerfile)
```



## Example

```bash
container-diff diff --type=apt --type=node --type=pip --json daemon://ubuntu:22.04 daemon://ubuntu:23.04 > diff.json
dockerfiler --input-file=diff.json --output-file=Dockerfile
```



## License

Project License can be found [here](LICENSE).



## Reference

- [container-diff](https://github.com/GoogleContainerTools/container-diff)
- [dockerfile-generator](https://www.startwithdocker.com/)
