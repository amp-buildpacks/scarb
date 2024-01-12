package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"scarb/scarb"
)

func main() {
	libcnb.Main(
		scarb.Detect{},
		scarb.Build{Logger: bard.NewLogger(os.Stdout)},
	)
}
