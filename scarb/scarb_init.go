package scarb

import (
	"bytes"
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sbom"
	"github.com/paketo-buildpacks/libpak/sherpa"
	"os"
	"path/filepath"
	"strings"
)

type ScarbInit struct {
	Version          string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
	Executor         effect.Executor
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

func (s ScarbInit) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	s.LayerContributor.Logger = s.Logger

	return s.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		file := filepath.Join(layer.Path, "bin", filepath.Base(artifact.Name()))

		s.Logger.Bodyf("Copying to %s", filepath.Dir(file))
		// 解压到PATH
		if err := crush.Extract(artifact, file, 0); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand %s\n%w", artifact.Name(), err)
		}
		if err := os.Chmod(filepath.Dir(file), 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		scarbPath := filepath.Join(file, fmt.Sprintf("scarb-%s-x86_64-unknown-linux-musl/bin/scarb", s.Version))

		if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", scarbPath)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
		}
		buf := &bytes.Buffer{}
		if err := s.Executor.Execute(effect.Execution{
			Command: scarbPath,
			Args:    []string{"-V"},
			Stdout:  buf,
			Stderr:  buf,
		}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("error executing '%s -V':\n Combined Output: %s: \n%w", file, buf.String(), err)
		}
		ver := strings.Split(strings.TrimSpace(buf.String()), " ")
		s.Logger.Bodyf("Checking %s version: %s", file, ver[1])
		sbomPath := layer.SBOMPath(libcnb.SyftJSON)
		dep := sbom.NewSyftDependency(layer.Path, []sbom.SyftArtifact{
			{
				ID:      "scarb-init",
				Name:    "Scarb",
				Version: s.Version,
				Type:    "UnknownPackage",
				FoundBy: "amp-buildpacks/scarb",
				Locations: []sbom.SyftLocation{
					{Path: "amp-buildpacks/scarb/scarb/scarb_init.go"},
				},
				Licenses: []string{"GNU"},
				CPEs:     []string{fmt.Sprintf("cpe:2.3:a:scarb:scarb:%s:*:*:*:*:*:*:*", s.Version)},
				PURL:     fmt.Sprintf("pkg:generic/scarb@%s", s.Version),
			},
		})
		s.Logger.Debugf("Writing Syft SBOM at %s: %+v", sbomPath, dep)
		if err := dep.WriteTo(sbomPath); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to write SBOM\n%w", err)
		}
		return layer, nil
	})
}

func (s ScarbInit) Name() string {
	return s.LayerContributor.LayerName()
}

func (s ScarbInit) BuildProcessTypes(runEnable string) ([]libcnb.Process, error) {
	processes := []libcnb.Process{}

	if runEnable == "true" {
		processes = append(processes, libcnb.Process{
			Type:      "web",
			Command:   "scarb cairo-run ",
			Arguments: []string{"--available-gas=200000000"},
			Default:   true,
			//WorkingDirectory: "",
		})
	}
	return processes, nil
}
