package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// --------------- UTILITY FUNCTIONS --------------- \\
func OpenInEditor(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filename)
	case "windows":
		cmd = exec.Command("notepad", filename)
	default:
		cmd = exec.Command("xdg-open", filename)
	}

	return cmd.Run()
}

// --------------- PROXY FUNCTIONS --------------- \\
func LoadProxyGroups() ([]ProxyGroup, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, "Popmart CLI", "proxies.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var groups []ProxyGroup
	err = json.Unmarshal(data, &groups)
	return groups, err
}

// --------------- SETTINGS FUNCTIONS --------------- \\
func ParseIntColor(hex string) int {
	var intValue int
	_, err := fmt.Sscanf(hex, "#%06x", &intValue)
	if err != nil {
		return 0
	}
	return intValue
}

func LoadSettings() (Settings, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Settings{}, err
	}

	path := filepath.Join(home, "Popmart CLI", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}

	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func SendWebhook(url string, data map[string]any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
