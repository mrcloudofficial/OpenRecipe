# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

# This Dockerfile builds on golang:alpine by building OpenRecipe from source
# using the current working directory.
#
# This produces a docker image that contains a working OpenRecipe binary along
# with all of its source code. This is not what produces the official releases
# in the "terraform" namespace on Dockerhub; those images include only
# the officially-released binary from releases.hashicorp.com and are
# built by the (closed-source) official release process.

FROM docker.mirror.hashicorp.services/golang:alpine
LABEL maintainer="HashiCorp OpenRecipe Team <terraform@hashicorp.com>"

RUN apk add --no-cache git bash openssh

ENV TF_DEV=true
ENV TF_RELEASE=1

WORKDIR $GOPATH/src/github.com/hashicorp/terraform
COPY . .
RUN /bin/bash ./scripts/build.sh

WORKDIR $GOPATH
ENTRYPOINT ["terraform"]