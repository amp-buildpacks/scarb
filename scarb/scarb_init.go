package scarb

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/sherpa"
	"os"
	"path/filepath"
)

type ScarbInit struct {
	Version          string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewScarbInit(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) ScarbInit {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Cache:  true,
		Launch: true,
		Build:  true,
	})
	return ScarbInit{
		Version:          dependency.Version,
		LayerContributor: contributor,
	}
}

func (r ScarbInit) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	r.LayerContributor.Logger = r.Logger

	return r.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		file := filepath.Join(layer.Path, "bin", filepath.Base(artifact.Name()))

		r.Logger.Bodyf("Copying to %s", filepath.Dir(file))
		// 解压到PATH
		if err := crush.Extract(artifact, file, 0); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand %s\n%w", artifact.Name(), err)
		}
		if err := os.Chmod(filepath.Dir(file), 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		scarbPath := filepath.Join(file, fmt.Sprintf("scarb-%s-x86_64-unknown-linux-musl/bin/scarb", r.Version))
		
		if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", scarbPath)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
		}
		return layer, nil
	})
}

func (r ScarbInit) Name() string {
	return r.LayerContributor.LayerName()
}
