package main

import (
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"os"
	"scarb/scarb"
)

func main() {
	libcnb.Main(
		scarb.Detect{},
		scarb.Build{Logger: bard.NewLogger(os.Stdout)},
	)
}
