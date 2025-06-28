package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(versionInfoURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	var data struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Version, nil
}

func Update() error {
	tmpFile := filepath.Join(os.TempDir(), "PopmartCLI_update.exe")

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find current exe: %w", err)
	}

	updaterPath := filepath.Join(filepath.Dir(exePath), "updater.exe")
	if _, err := os.Stat(updaterPath); os.IsNotExist(err) {
		return fmt.Errorf("updater.exe not found at %s", updaterPath)
	}

	cmd := exec.Command("cmd", "/C", "start", "", updaterPath, exePath, tmpFile)
	return cmd.Start()
}
