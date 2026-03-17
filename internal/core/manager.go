package core

import "github.com/nyable/nyaru/internal/models"

// PackageManager defines the operations supported by the package manager
type PackageManager interface {
	Search(query string) ([]models.AppInfo, error)
	List() ([]models.AppInfo, error)
	Status() ([]models.AppInfo, error)
	Install(app string) error
	Uninstall(app string) error
	Info(app string) (string, error)
	Update(apps ...string) error
	BucketList() ([]models.BucketResult, error)
	BucketAdd(name, source string) error
	BucketRemove(name string) error
	CacheList() ([]models.CacheResult, error)
	CacheRemove(names ...string) error
}


var currentManager PackageManager

// GetManager returns the appropriate package manager based on the requested mode.
// If mode is "sfsu", it returns an SfsuManager.
// If mode is "scoop", it returns a NativeScoopManager.
func GetManager(mode string) PackageManager {
	if mode == "sfsu" {
		if currentManager == nil {
			currentManager = NewSfsuManager()
		} else {
			if _, ok := currentManager.(*SfsuManager); !ok {
				currentManager = NewSfsuManager()
			}
		}
	} else {
		if currentManager == nil {
			currentManager = NewNativeScoopManager()
		} else {
			if _, ok := currentManager.(*NativeScoopManager); !ok {
				currentManager = NewNativeScoopManager()
			}
		}
	}
	return currentManager
}
