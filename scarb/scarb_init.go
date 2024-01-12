/*
 * Copyright 2018-2020 the original author or authors.
 *
 * COPY FROM amp-buildpacks/scarb.git
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
		Executor:         effect.NewExecutor(),
		Version:          dependency.Version,
		LayerContributor: contributor,
	}
}

func (s ScarbInit) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	s.LayerContributor.Logger = s.Logger
	return s.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		bin := filepath.Join(layer.Path, "bin")

		s.Logger.Bodyf("Expanding %s to %s", artifact.Name(), layer.Path)
		if err := crush.Extract(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand %s\n%w", artifact.Name(), err)
		}

		s.Logger.Bodyf("Setting %s as executable", bin)
		file := filepath.Join(bin, "scarb")
		if err := os.Chmod(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		s.Logger.Bodyf("Setting %s in PATH", layer.Path)
		if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", bin)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
		}

		buf := &bytes.Buffer{}
		if err := s.Executor.Execute(effect.Execution{
			Command: "scarb",
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
				ID:      "scarb_init",
				Name:    "scarb",
				Version: ver[1],
				Type:    "UnknownPackage",
				FoundBy: "amp-buildpacks/scarb",
				Locations: []sbom.SyftLocation{
					{Path: "amp-buildpacks/scarb/scarb/scarb_init.go"},
				},
				Licenses: []string{"MIT"},
				CPEs:     []string{fmt.Sprintf("cpe:2.3:a:scarb:scarb:%s:*:*:*:*:*:*:*", ver[1])},
				PURL:     fmt.Sprintf("pkg:generic/scarb@%s", ver[1]),
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
		})
	}
	return processes, nil
}
