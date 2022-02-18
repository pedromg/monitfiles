#!/usr/bin/env sh
#

# check the `go test` result to create or not the build
# Only create if the tests pass:
#  - 0 means tests passed
#  - 1 is a test fail
#  - 2 is a compile error
test=$(go test -v ./...)
if [ $? -ne 0 ]; then
	echo $test
	exit
fi

build=$(go build)
if [ $? -ne 0 ]; then
	echo $build
	exit
fi

git_hash=$(git rev-parse HEAD)
base_hash=$(git rev-list --max-count=1 HEAD -- VERSION.txt)
change_count=$(git rev-list --count HEAD)
ver='0.'$change_count
echo $base_hash > go.toolchain.rev
echo $ver > VERSION.txt
git tag $ver
go version > go.toolchain.ver
