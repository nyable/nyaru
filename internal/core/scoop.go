package core

import (
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/utils"
)

type NativeScoopManager struct{}

func NewNativeScoopManager() *NativeScoopManager {
	return &NativeScoopManager{}
}

func (m *NativeScoopManager) Search(query string) ([]models.AppInfo, error) {
	strOutput, _, _, err := utils.RunHookedCommand("search", query)
	if err != nil {
		return nil, err
	}
	// Parse powershell output into models.AppInfo
	return utils.PsDirtyJSONToStructList[models.AppInfo](strOutput)
}

func (m *NativeScoopManager) List() ([]models.AppInfo, error) {
	strOutput, _, _, err := utils.RunHookedCommand("list", "")
	if err != nil {
		return nil, err
	}
	return utils.PsDirtyJSONToStructList[models.AppInfo](strOutput)
}

func (m *NativeScoopManager) Status() ([]models.AppInfo, error) {
	strOutput, _, _, err := utils.RunHookedCommand("status", "")
	if err != nil {
		return nil, err
	}
	return utils.PsDirtyJSONToStructList[models.AppInfo](strOutput)
}

func (m *NativeScoopManager) Install(app string) error {
	return runInteractiveCommand("scoop", "install", app)
}

func (m *NativeScoopManager) Uninstall(app string) error {
	return runInteractiveCommand("scoop", "uninstall", app)
}

func (m *NativeScoopManager) Info(app string) (string, error) {
	return runCommandOutput("scoop", "info", app)
}

func (m *NativeScoopManager) Update(apps ...string) error {
	args := append([]string{"update"}, apps...)
	return runInteractiveCommand("scoop", args...)
}

func (m *NativeScoopManager) BucketList() ([]models.BucketResult, error) {
	strOutput, _, _, err := utils.RunHookedCommand("bucket list", "")
	if err != nil {
		return nil, err
	}
	return utils.PsDirtyJSONToStructList[models.BucketResult](strOutput)
}

func (m *NativeScoopManager) BucketAdd(name, source string) error {
	return runInteractiveCommand("scoop", "bucket", "add", name, source)
}

func (m *NativeScoopManager) Hold(app string) error {
	return runInteractiveCommand("scoop", "hold", app)
}

func (m *NativeScoopManager) Unhold(app string) error {
	return runInteractiveCommand("scoop", "unhold", app)
}

func (m *NativeScoopManager) BucketRemove(name string) error {
	return runInteractiveCommand("scoop", "bucket", "rm", name)
}

func (m *NativeScoopManager) CacheList() ([]models.CacheResult, error) {
	// Use powershell to ensure scoop (which can be a function/shim) is called correctly
	output, _, err := utils.RunWithPowerShellCombined("powershell", "-Command", "scoop cache show")
	if err != nil {
		if output == "" {
			return nil, err
		}
	}
	results := utils.ParseScoopCacheOutput(output)
	// if len(results) == 0 && output != "" {
	// fmt.Printf("DEBUG: Raw Scoop Output:\n%s\n", output)
	// }
	return results, nil
}

func (m *NativeScoopManager) CacheRemove(names ...string) error {
	args := append([]string{"cache", "rm"}, names...)
	return runInteractiveCommand("scoop", args...)
}
