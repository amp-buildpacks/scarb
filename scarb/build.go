/*
 * Copyright 2018-2020 the original author or authors.
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
	"fmt"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"log"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()
	config, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	build_scarb, _ := config.Resolve("BP_ENABLE_SCARB_PROCESS")
	dependency, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, err
	}
	build_dependency, _ := dependency.Resolve("scarb-init", "*")
	log.Println("scarb dependency  = %+v", build_dependency)

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	scarbInit := NewScarbInit(build_dependency, dc)
	result.Processes, err = scarbInit.BuildProcessTypes(build_scarb)
	result.Layers = append(result.Layers, scarbInit)
	return result, nil
}
