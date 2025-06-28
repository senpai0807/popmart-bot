package api

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

//go:embed js/vm.cjs
var vmCJS string

//go:embed js/fingerprint.js
var fingerprintJS string

//go:embed js/decode.js
var decodeJS string

//go:embed js/session.js
var sessionJS string

func GetTimestamp() int {
	return int(time.Now().Unix())
}

func GenerateSignature(timestamp int) string {
	clientId := "nw3b089qrgw9m7b7i"
	input := strconv.Itoa(timestamp) + "," + clientId

	hash := md5.Sum([]byte(input))
	signature := hex.EncodeToString(hash[:])

	xSign := signature + "," + strconv.Itoa(timestamp)
	return xSign
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		return ""
	}
}

func sortMapRecursively(input any) any {
	switch val := input.(type) {
	case []any:
		sortedArray := make([]any, len(val))
		for i, v := range val {
			sortedArray[i] = sortMapRecursively(v)
		}
		return sortedArray

	case map[string]any:
		sorted := make(map[string]any)
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sorted[k] = sortMapRecursively(val[k])
		}
		return sorted

	default:
		return val
	}
}

func GenerateSData(data map[string]any, timestamp int, method string, secretKey string) map[string]any {
	if secretKey == "" {
		secretKey = "W_ak^moHpMla"
	}

	sortedData := sortMapRecursively(data).(map[string]any)
	filteredData := make(map[string]any)

	keys := make([]string, 0, len(sortedData))
	for k := range sortedData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := sortedData[k]
		if method == "get" {
			if v != nil {
				switch val := v.(type) {
				case string:
					if val != "" {
						filteredData[k] = val
					}
				default:
					filteredData[k] = toString(val)
				}
			}
		} else {
			filteredData[k] = v
		}
	}

	jsonBytes, _ := json.Marshal(filteredData)
	toHash := string(jsonBytes) + secretKey + strconv.Itoa(timestamp)

	hash := md5.Sum([]byte(toHash))
	s := hex.EncodeToString(hash[:])

	return map[string]any{
		"s": s,
		"t": timestamp,
	}
}

func CreateTLSClient(proxyUrl string) (tls_client.HttpClient, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithForceHttp1(),
		tls_client.WithCookieJar(jar),
		tls_client.WithProxyUrl(proxyUrl),
		tls_client.WithTimeoutSeconds(120),
		tls_client.WithInsecureSkipVerify(),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithClientProfile(profiles.Chrome_133),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func RunJS() (string, error) {
	tmpDir, err := os.MkdirTemp("", "popmartjs")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fpPath := filepath.Join(tmpDir, "fingerprint.js")
	vmPath := filepath.Join(tmpDir, "vm.cjs")

	if err := os.WriteFile(fpPath, []byte(fingerprintJS), 0644); err != nil {
		return "", fmt.Errorf("failed to write fingerprint.js: %w", err)
	}
	if err := os.WriteFile(vmPath, []byte(vmCJS), 0644); err != nil {
		return "", fmt.Errorf("failed to write vm.cjs: %w", err)
	}

	cmd := exec.Command("node", fpPath)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("JS execution failed: %v\nStderr: %s", err, stderr.String())
	}

	result := strings.TrimSpace(out.String())
	return result, nil
}

func RunDecode(input TdResp) (DecodeResp, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return DecodeResp{}, fmt.Errorf("failed to encode input: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "popmartjs2")
	if err != nil {
		return DecodeResp{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	decodePath := filepath.Join(tmpDir, "decode.js")
	vmPath := filepath.Join(tmpDir, "vm.cjs")

	if err := os.WriteFile(decodePath, []byte(decodeJS), 0644); err != nil {
		return DecodeResp{}, fmt.Errorf("failed to write decode.js: %w", err)
	}

	if err := os.WriteFile(vmPath, []byte(vmCJS), 0644); err != nil {
		return DecodeResp{}, fmt.Errorf("failed to write vm.cjs: %w", err)
	}

	cmd := exec.Command("node", decodePath, string(jsonData))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return DecodeResp{}, fmt.Errorf("JS failed: %v\nStderr: %s", err, stderr.String())
	}

	var result DecodeResp
	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return DecodeResp{}, fmt.Errorf("failed to parse output: %w", err)
	}

	return result, nil
}

func RunSession(tokenId, path, body string) (SessionResp, error) {
	tmpDir, err := os.MkdirTemp("", "popmartjs3")
	if err != nil {
		return SessionResp{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	sessionPath := filepath.Join(tmpDir, "session.js")
	vmPath := filepath.Join(tmpDir, "vm.cjs")

	if err := os.WriteFile(sessionPath, []byte(sessionJS), 0644); err != nil {
		return SessionResp{}, fmt.Errorf("failed to write session.js: %w", err)
	}
	if err := os.WriteFile(vmPath, []byte(vmCJS), 0644); err != nil {
		return SessionResp{}, fmt.Errorf("failed to write vm.cjs: %w", err)
	}

	cmd := exec.Command("node", sessionPath, tokenId, path, body)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return SessionResp{}, fmt.Errorf("JS execution failed: %v\nStderr: %s", err, stderr.String())
	}

	var result SessionResp
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return SessionResp{}, fmt.Errorf("failed to parse JS output: %w", err)
	}
	return result, nil
}
