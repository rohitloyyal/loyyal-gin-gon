#!/bin/bash

# Usage
# ./dockerbuild.sh cmd/servicename

fail() { echo >&2 $*; exit 1; }

# setup
set -ex

# work from top of project, this allows compiling with vendor dir
# to temporarily modify vendored code with debug statements
# by prefacing the make command with GO111MODULE=off
cd $(dirname ${0})/..

[[ -n $VERSION ]] || VERSION=$(git describe --tags 2>/dev/null || echo dev(git rev-parse --short HEAD))

DIR=$1
[[ -d $DIR ]] || fail "directory '$DIR' not specified or does not exist"
DIR=$(cd $DIR; pwd) # abs path
FILE=$(basename $DIR)

go build -v -ldflags "-s -w -X main.version=$VERSION" -o $DIR/$FILE $DIR

# 2nd arg is optional top level dir
cd ${2:-$DIR}
aws ecr get-login-password --region me-south-1 | docker login --username AWS --password-stdin 827830277284.dkr.ecr.me-south-1.amazonaws.com
docker build -t 827830277284.dkr.ecr.me-south-1.amazonaws.com/$FILE:$VERSION -f $DIR/Dockerfile .
docker push 827830277284.dkr.ecr.me-south-1.amazonaws.com/$FILE:$VERSION
# cleanup
rm $DIR/$FILE