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
	"github.com/buildpacks/libcnb"
	"os"
	"path/filepath"
)

type Detect struct {
}

const PlanEntryScarb = "scarb"

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	found, err := d.CairoProject(context.Application.Path)
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to detect scarb requirements\n%w", err)

	}
	if !found {
		return libcnb.DetectResult{Pass: false}, nil
	}
	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: PlanEntryScarb},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: PlanEntryScarb},
				},
			},
		},
	}, nil
}

func (d Detect) CairoProject(appDir string) (bool, error) {
	_, err := os.Stat(filepath.Join(appDir, "Scarb.toml"))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("unable to determine if Scarb.toml exists\n%w", err)
	}
	return true, nil
}
