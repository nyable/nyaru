package core

import (
	"encoding/json"
	"fmt"

	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/tui"
	"github.com/nyable/nyaru/internal/utils"
)

type SfsuManager struct{}

func NewSfsuManager() *SfsuManager {
	return &SfsuManager{}
}

func (m *SfsuManager) runSfsuQueryList(action, query string) ([]models.AppInfo, error) {
	strOutput, _, isSfsu, err := utils.RunHookedCommand(action, query)
	if err != nil {
		return nil, err
	}
	if !isSfsu {
		return nil, fmt.Errorf("sfsu fallback to scoop unexpectedly")
	}
	return utils.StandardJSONToStructList[models.AppInfo](strOutput)
}

func (m *SfsuManager) Search(query string) ([]models.AppInfo, error) {
	strOutput, _, isSfsu, err := utils.RunHookedCommand("search", query)
	if err != nil {
		return nil, err
	}
	if !isSfsu {
		return nil, fmt.Errorf("sfsu fallback to scoop unexpectedly")
	}

	// sfsu search returns a map: bucket_name -> []AppInfo
	var result map[string][]models.AppInfo
	if err := json.Unmarshal([]byte(strOutput), &result); err != nil {
		return nil, fmt.Errorf("failed to parse sfsu search output: %w", err)
	}

	var flatList []models.AppInfo
	for bucket, apps := range result {
		for i := range apps {
			apps[i].Bucket = bucket
		}
		flatList = append(flatList, apps...)
	}
	return flatList, nil
}

func (m *SfsuManager) List() ([]models.AppInfo, error) {
	return m.runSfsuQueryList("list", "")
}

func (m *SfsuManager) Status() ([]models.AppInfo, error) {
	strOutput, _, isSfsu, err := utils.RunHookedCommand("status", "")
	if err != nil {
		return nil, err
	}
	if !isSfsu {
		return nil, fmt.Errorf("sfsu fallback to scoop unexpectedly")
	}

	// sfsu status returns: {"scoop":false,"buckets":[],"packages":[]}
	var result struct {
		Packages []struct {
			Name      string `json:"name"`
			Current   string `json:"current"`
			Available string `json:"available"`
		} `json:"packages"`
	}

	if err := json.Unmarshal([]byte(strOutput), &result); err != nil {
		return nil, fmt.Errorf("failed to parse sfsu status output: %w", err)
	}

	var list []models.AppInfo
	for _, p := range result.Packages {
		list = append(list, models.AppInfo{
			Name:      p.Name,
			Version:   p.Available, // Show available version as the "version" for update
			Installed: true,
		})
	}
	return list, nil
}

func (m *SfsuManager) Install(app string) error {
	return runInteractiveCommand("scoop", "install", app)
}

func (m *SfsuManager) Uninstall(app string) error {
	return runInteractiveCommand("scoop", "uninstall", app)
}

func (m *SfsuManager) Info(app string) (string, error) {
	return runCommandOutput("sfsu", "info", app, "--json")
}

func (m *SfsuManager) Update(apps ...string) error {
	if len(apps) > 0 {
		args := append([]string{"update"}, apps...)
		return runInteractiveCommand("scoop", args...)
	}
	// No args: Update both scoop and sfsu index
	tui.PrintInfo("正在更新 Scoop 核心...")
	if err := runInteractiveCommand("scoop", "update"); err != nil {
		return err
	}
	tui.PrintInfo("正在更新 sfsu 索引...")
	return runInteractiveCommand("sfsu", "update")
}

func (m *SfsuManager) BucketList() ([]models.BucketResult, error) {
	strOutput, _, isSfsu, err := utils.RunHookedCommand("bucket list", "")
	if err != nil {
		return nil, err
	}
	if isSfsu {
		return utils.StandardJSONToStructList[models.BucketResult](strOutput)
	}
	return utils.PsDirtyJSONToStructList[models.BucketResult](strOutput)
}

func (m *SfsuManager) BucketAdd(name, source string) error {
	return runInteractiveCommand("scoop", "bucket", "add", name, source)
}

func (m *SfsuManager) BucketRemove(name string) error {
	return runInteractiveCommand("scoop", "bucket", "rm", name)
}

func (m *SfsuManager) Hold(app string) error {
	return runInteractiveCommand("scoop", "hold", app)
}

func (m *SfsuManager) Unhold(app string) error {
	return runInteractiveCommand("scoop", "unhold", app)
}

func (m *SfsuManager) CacheList() ([]models.CacheResult, error) {
	// sfsu doesn't support cache show --json yet, fallback to native scoop parsing
	output, _, err := utils.RunWithPowerShellCombined("powershell", "-Command", "scoop cache show")
	if err != nil {
		if output == "" {
			return nil, err
		}
	}
	results := utils.ParseScoopCacheOutput(output)
	// if len(results) == 0 && output != "" {
	// 	fmt.Printf("DEBUG: Raw Scoop Output:\n%s\n", output)
	// }
	return results, nil
}

func (m *SfsuManager) CacheRemove(names ...string) error {
	args := append([]string{"cache", "rm"}, names...)
	return runInteractiveCommand("scoop", args...)
}
