#!/usr/bin/env bash

set -eu -o pipefail

BUILDKIT_VERSION=0.13.0

curl -sfL https://github.com/moby/buildkit/releases/download/v${BUILDKIT_VERSION}/buildkit-v${BUILDKIT_VERSION}.linux-amd64.tar.gz -o buildkit.tar.gz
sudo tar -zxvf buildkit.tar.gz -C /usr/local/
rm buildkit.tar.gz
