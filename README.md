# copatcher

[![Build Status](https://github.com/craftslab/copatcher/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/craftslab/copatcher/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/craftslab/copatcher/branch/main/graph/badge.svg?token=7PMQALLZLY)](https://codecov.io/gh/craftslab/copatcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/copatcher)](https://goreportcard.com/report/github.com/craftslab/copatcher)
[![License](https://img.shields.io/github/license/craftslab/copatcher.svg)](https://github.com/craftslab/copatcher/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/copatcher.svg)](https://github.com/craftslab/copatcher/tags)



## Introduction

*copatcher* is a container patcher written in Go.



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
./bin/copatcher --container-diff=diff.json --output-file=Dockerfile
```



## Docker

```bash
version=latest make docker
docker run ghcr.io/craftslab/copatcher:latest
```



## Usage

```
usage: copatcher --container-diff=CONTAINER-DIFF --output-file=OUTPUT-FILE [<flags>]

Container patcher


Flags:
  --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
  --[no-]version             Show application version.
  --container-diff=CONTAINER-DIFF
                             Container difference (.json)
  --output-file=OUTPUT-FILE  Output file (Dockerfile)
```



## Example

```bash
container-diff diff --type=apt --type=node --type=pip --json daemon://ubuntu:22.04 daemon://ubuntu:23.04 > diff.json
copatcher --container-diff=diff.json --output-file=Dockerfile
```



## License

Project License can be found [here](LICENSE).



## Reference

- [container-diff](https://github.com/GoogleContainerTools/container-diff)
- [copacetic](https://project-copacetic.github.io/copacetic/website/)
- [dockerfile-generator](https://www.startwithdocker.com/)
