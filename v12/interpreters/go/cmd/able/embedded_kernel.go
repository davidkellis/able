package main

import "embed"

//go:embed embedded/kernel/src/kernel.able
//go:embed embedded/kernel/package.yml
var embeddedKernelFS embed.FS
