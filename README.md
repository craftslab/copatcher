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



## Test

```bash
container-diff diff --type=apt --type=node --type=pip --json \
  daemon://ubuntu:22.04 daemon://ubuntu:23.04 > diff.json
```



## Run

```bash
version=latest make build
./bin/dockerfiler --input-image1=image1 --input-image2=image2 --output-file=/path/to/Dockerfile
```



## Usage

```
TBD
```



## Example

```bash
dockerfiler --input-image1=daemon://ubuntu:22.04 --input-image2=daemon://ubuntu:23.04 --output-file=Dockerfile
```



## License

Project License can be found [here](LICENSE).



## Reference

- [container-diff](https://github.com/GoogleContainerTools/container-diff)
- [dockerfile-generator](https://www.startwithdocker.com/)
