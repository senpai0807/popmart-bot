package tasks

import (
	"os"
	"path/filepath"

	backend "popmart/src/backend"
)

var (
	ProxyGroups   []backend.ProxyGroup
	AccountGroups []backend.AccountGroup
	TasksPath     string
	ProxiesPath   string
	ProfilesPath  string
	AccountsPath  string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	TasksPath = filepath.Join(home, "Popmart CLI", "tasks.csv")
	ProxiesPath = filepath.Join(home, "Popmart CLI", "proxies.json")
	ProfilesPath = filepath.Join(home, "Popmart CLI", "profiles.csv")
	AccountsPath = filepath.Join(home, "Popmart CLI", "accounts.json")
}
