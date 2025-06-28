package proxies

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	backend "popmart/src/backend"
	helpers "popmart/src/middleware/helpers"

	http "github.com/bogdanfinn/fhttp"
	"github.com/google/uuid"
)

func AddProxyGroup(logger *helpers.ColorizedLogger, groupName string) error {
	tmpFile, err := os.CreateTemp("", "proxies_*.txt")
	if err != nil {
		logger.Error("Failed To Create Temporary Text Document")
		return err
	}

	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	defer os.Remove(tmpFilePath)

	if err := backend.OpenInEditor(tmpFilePath); err != nil {
		logger.Error("Failed To Open Text Editor")
		return err
	}

	file, err := os.Open(tmpFilePath)
	if err != nil {
		logger.Error("Failed To Read Text Document After Editing")
		return err
	}
	defer file.Close()

	var proxies []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 4)
		if len(parts) != 4 {
			logger.Warn(fmt.Sprintf("Line %d Skipped: Expected Format host:port:user:pass", lineNum))
			continue
		}

		host := strings.TrimSpace(parts[0])
		port := strings.TrimSpace(parts[1])
		user := strings.TrimSpace(parts[2])
		pass := strings.TrimSpace(parts[3])

		if host == "" || port == "" || user == "" || pass == "" {
			logger.Warn(fmt.Sprintf("Line %d Skipped: One Or More Fields Are Empty", lineNum))
			continue
		}

		if net.ParseIP(host) == nil {
			logger.Warn(fmt.Sprintf("Line %d Skipped: Invalid IP Address: %s", lineNum, host))
			continue
		}

		proxies = append(proxies, fmt.Sprintf("%s:%s:%s:%s", host, port, user, pass))
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Failed To Read Lines From Text File")
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed To Resolve User Home Directory")
		return err
	}

	path := filepath.Join(home, "Popmart CLI", "proxies.json")

	var proxyGroups []backend.ProxyGroup

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &proxyGroups)
	}

	proxyGroups = append(proxyGroups, backend.ProxyGroup{
		Name:    groupName,
		ID:      uuid.New().String(),
		Proxies: proxies,
	})

	updated, err := json.MarshalIndent(proxyGroups, "", "  ")
	if err != nil {
		logger.Error("Failed To Marshal Proxies JSON")
		return err
	}

	err = os.WriteFile(path, updated, 0644)
	if err != nil {
		logger.Error("Failed To Write To Proxies File")
		return err
	}

	return nil
}

func TestProxyGroup(logger *helpers.ColorizedLogger, group backend.ProxyGroup) {
	logger.Info(fmt.Sprintf("Testing Proxies In Group: %s", group.Name))
	for _, proxyStr := range group.Proxies {
		start := time.Now()

		parts := strings.Split(proxyStr, ":")
		if len(parts) != 4 {
			logger.Warn(fmt.Sprintf("Invalid Proxy Format Skipped: %s", proxyStr))
			continue
		}

		client, err := helpers.CreateTLSClient(fmt.Sprintf("http://%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1]))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Create Request Client: %v", err))
			continue
		}

		req, err := http.NewRequest("GET", "https://www.popmart.com/us", nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Initialize Request: %v", err))
			continue
		}

		req.Header = http.Header{
			"sec-ch-ua":                 {`"Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"`},
			"sec-ch-ua-mobile":          {"?0"},
			"sec-ch-ua-platform":        {`"Windows"`},
			"upgrade-insecure-requests": {"1"},
			"user-agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"},
			"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
			"sec-fetch-site":            {"none"},
			"sec-fetch-mode":            {"navigate"},
			"sec-fetch-user":            {"?1"},
			"sec-fetch-dest":            {"document"},
			"accept-encoding":           {"gzip, deflate, br, zstd"},
			"accept-language":           {"en-US,en;q=0.9"},
			"priority":                  {"u=0, i"},
			"Header-Order:": {
				"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept",
				"sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language", "priority",
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Execute Request: %v", err))
			continue
		}
		resp.Body.Close()

		elapsed := time.Since(start)
		logger.Silly(fmt.Sprintf("%s - Speed [%dms]", proxyStr, elapsed.Milliseconds()))
	}
}
