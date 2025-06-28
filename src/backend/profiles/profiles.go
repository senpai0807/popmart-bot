package profiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenProfilesCSV() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve user home directory: %w", err)
	}

	profilesPath := filepath.Join(home, "Popmart CLI", "profiles.csv")

	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		return fmt.Errorf("profiles.csv does not exist at %s", profilesPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", profilesPath)
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "", profilesPath)
	default:
		cmd = exec.Command("xdg-open", profilesPath)
	}

	return cmd.Run()
}
