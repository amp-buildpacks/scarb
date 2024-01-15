// Copyright (c) The Amphitheatre Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scarb

import (
	"fmt"
	"log"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()
	config, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	buildScarb, _ := config.Resolve("BP_ENABLE_SCARB_PROCESS")
	dependency, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, err
	}
	libc, _ := config.Resolve("BP_SCARB_LIBC")

	version, _ := config.Resolve("BP_LEO_VERSION")
	buildDependency, _ := dependency.Resolve(fmt.Sprintf("scarb-%s", libc), version)
	log.Printf("scarb dependency  = %+v", buildDependency)

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	scarb := NewScarb(buildDependency, dc)
	scarb.Logger = b.Logger
	result.Processes, err = scarb.BuildProcessTypes(buildScarb)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to build list of process types\n%w", err)
	}
	result.Layers = append(result.Layers, scarb)
	return result, nil
}
