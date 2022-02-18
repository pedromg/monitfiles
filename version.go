package main

import _ "embed"

//go:embed VERSION.txt
var version []byte

// goToolChainRev holds the last git hash that the release was built
//go:embed go.toolchain.rev
var goToolChainRev string

// goToolChainVer holds the Go version of the build
//go:embed go.toolchain.ver
var goToolChainVer string
