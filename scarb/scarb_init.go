package scarb

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
	"os"
	"path/filepath"
)

type ScarbInit struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewScarbInit(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) ScarbInit {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Cache: true,
	})
	return ScarbInit{
		LayerContributor: contributor,
	}
}

func (r ScarbInit) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	r.LayerContributor.Logger = r.Logger

	if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", filepath.Join(layer.Path, "bin"))); err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
	}

	return r.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		file := filepath.Join(layer.Path, "bin", filepath.Base(artifact.Name()))

		r.Logger.Bodyf("Copying to %s", filepath.Dir(file))

		if err := sherpa.CopyFile(artifact, file); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy %s to %s\n%w", artifact.Name(), file, err)
		}

		if err := os.Chmod(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		return layer, nil
	})
}

func (r ScarbInit) Name() string {
	return r.LayerContributor.LayerName()
}
