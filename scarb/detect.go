package scarb

import (
	"fmt"
	"github.com/buildpacks/libcnb"
	"os"
	"path/filepath"
)

type Detect struct {
}

const PlanEntryCairo = "cairo"

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
					{Name: PlanEntryCairo},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: PlanEntryCairo},
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

	_, err = os.Stat(filepath.Join(appDir, "Scarb.lock"))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("unable to determine if Scarb.lock exists\n%w", err)
	}

	return true, nil
}
