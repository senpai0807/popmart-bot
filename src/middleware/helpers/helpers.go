/*
- LOGGER FUNCTION
- UPDATER FUNCTIONS
- INITIALIZE FILES FUNCTION
- TASK FUNCTIONS
- REQUEST CLIENT
*/
package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	update "popmart/src/middleware/helpers/update"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/fatih/color"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

// ---------------------- LOGGER FUNCTION ---------------------- \\
func FormatDate(t time.Time) string {
	return t.Format("03:04:05 PM - 01/02/2006")
}

var colorCodes = map[string]func(a ...any) string{
	"info":    color.New(color.FgBlue).SprintFunc(),
	"verbose": color.New(color.FgCyan).SprintFunc(),
	"warn":    color.New(color.FgYellow).SprintFunc(),
	"error":   color.New(color.FgRed).SprintFunc(),
	"http":    color.New(color.FgMagenta).SprintFunc(),
	"silly":   color.New(color.FgGreen).SprintFunc(),
}

func (l *ColorizedLogger) log(level, message string) {
	timestamp := FormatDate(time.Now())
	colorFunc, exists := colorCodes[level]
	if !exists {
		colorFunc = color.New(color.Reset).SprintFunc()
	}

	var logMessage string
	if l.useColor {
		logMessage = fmt.Sprintf("%s: %s\n", colorFunc(timestamp), colorFunc(message))
	} else {
		logMessage = fmt.Sprintf("[%s]: %s\n", timestamp, message)
	}

	os.Stdout.WriteString(logMessage)
}

func NewColorizedLogger(useColor bool) *ColorizedLogger {
	return &ColorizedLogger{useColor: useColor}
}

func (l *ColorizedLogger) Info(message string)    { l.log("info", message) }
func (l *ColorizedLogger) Verbose(message string) { l.log("verbose", message) }
func (l *ColorizedLogger) Warn(message string)    { l.log("warn", message) }
func (l *ColorizedLogger) HTTP(message string)    { l.log("http", message) }
func (l *ColorizedLogger) Silly(message string)   { l.log("silly", message) }
func (l *ColorizedLogger) Error(message string)   { l.log("error", message) }

// ---------------------- UPDATER FUNCTIONS ---------------------- \\
func downloadUpdater(dest string) error {
	const updaterURL = "" // URL To Download updater.exe From CDN - I Used Digital Oceans

	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create updater file: %w", err)
	}
	defer outFile.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(updaterURL)
	if err != nil {
		return fmt.Errorf("failed to download updater: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("updater download failed: HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save updater file: %w", err)
	}

	return nil
}

func EnsureUpdaterExists(logger *ColorizedLogger) {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed To Get Users Home Directory: " + err.Error())
		return
	}
	baseDir := filepath.Join(home, "Popmart CLI")
	updaterPath := filepath.Join(baseDir, "updater.exe")

	if _, err := os.Stat(updaterPath); os.IsNotExist(err) {
		logger.Warn("updater.exe not found, downloading...")
		err := downloadUpdater(updaterPath)
		if err != nil {
			logger.Error("Failed To Download Updater: " + err.Error())
		} else {
			logger.Info("updater.exe downloaded successfully ✅")
		}
	}
}

// ---------------------- INITIALIZE FILES FUNCTION ---------------------- \\
func createTasksCSV(path string) {
	headers := `Task Group,Site,Mode,Input,Size,Proxy Group,Profile Group,Profile,Account Group,Quantity,Delay,Payment Method`
	os.WriteFile(path, []byte(headers), 0644)
}

func createProfilesCSV(path string) {
	headers := `Profile Group Name,Profile Name,Email,Name,Phone,Address 1,Address 2,City,Post Code,Country,State,Card Number,Expiration Month,Expiration Year,Security Code`
	os.WriteFile(path, []byte(headers), 0644)
}

func createEmptyJSONArray(path string) {
	empty := []byte("[]")
	os.WriteFile(path, empty, 0644)
}

func createSettingsJSON(path string) {
	settings := map[string]string{
		"webhookUrl": "",
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(path, data, 0644)
}

func InitFileSystem(logger *ColorizedLogger) {
	logger.Info("Initializing Popmart Engine")
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed To Get Users Home Directory: " + err.Error())
		os.Exit(1)
	}

	baseDir := filepath.Join(home, "Popmart CLI")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		err = os.Mkdir(baseDir, 0755)
		if err != nil {
			logger.Error("Failed To Create Popmart CLI Directory: " + err.Error())
			os.Exit(1)
		}
	}

	files := map[string]func(string){
		"tasks.csv":     createTasksCSV,
		"profiles.csv":  createProfilesCSV,
		"proxies.json":  createEmptyJSONArray,
		"accounts.json": createEmptyJSONArray,
		"sessions.json": createEmptyJSONArray,
		"settings.json": createSettingsJSON,
	}

	for filename, createFunc := range files {
		fullPath := filepath.Join(baseDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			createFunc(fullPath)
		}
	}
	EnsureUpdaterExists(logger)
}

// ---------------------- TASK FUNCTIONS ---------------------- \\
func Delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func IncrementCarted() {
	Mu.Lock()
	defer Mu.Unlock()
	Carted++
	updateTitle()
}

func IncrementCheckedOut() {
	Mu.Lock()
	defer Mu.Unlock()
	CheckedOut++
	updateTitle()
}

func updateTitle() {
	var ver update.VersionInfo
	if err := json.Unmarshal(update.VersionData, &ver); err != nil {
		os.Exit(1)
	}

	title := fmt.Sprintf("v%s | Carted: %d | Secured: %d", ver.Version, Carted, CheckedOut)
	fmt.Printf("\033]0;%s\007", title)
}

func initializeAudioCtx() {
	ctxInit.Do(func() {
		audioCtx, initErr = oto.NewContext(44100, 2, 2, 8192)
	})
}

func PlaySound(data []byte) error {
	decoder, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error decoding MP3: %w", err)
	}

	initializeAudioCtx()
	if initErr != nil {
		return fmt.Errorf("error initializing Oto context: %w", initErr)
	}

	player := audioCtx.NewPlayer()
	defer player.Close()

	if _, err := io.Copy(player, decoder); err != nil {
		return fmt.Errorf("error playing MP3: %w", err)
	}

	return nil
}

// ---------------------- REQUEST CLIENT ---------------------- \\
func CreateTLSClient(proxyUrl string) (tls_client.HttpClient, error) {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithCookieJar(jar),
		tls_client.WithTimeoutSeconds(120),
		tls_client.WithInsecureSkipVerify(),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithClientProfile(profiles.Chrome_133),
	}

	if proxyUrl != "" {
		options = append(options, tls_client.WithProxyUrl(proxyUrl))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ---------------------- SESSION FUNCTIONS ---------------------- \\
func FetchSession(logger *ColorizedLogger, taskId, accountEmail string) (UserData, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Get User Home Directory: %s", taskId, err.Error()))
		return UserData{}, err
	}

	sessionPath := filepath.Join(home, "Popmart CLI", "sessions.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Read Sessions File: %s", taskId, err.Error()))
		return UserData{}, err
	}

	var sessions []Session
	if len(data) > 0 {
		if err := json.Unmarshal(data, &sessions); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Parse Sessions File: %s", taskId, err.Error()))
			return UserData{}, err
		}
	}

	for _, session := range sessions {
		if session.AccountEmail == accountEmail {
			return UserData{
				AccessToken: session.AccessToken,
				GID:         session.GID,
			}, nil
		}
	}

	return UserData{}, fmt.Errorf("no session was found")
}

func SaveSession(logger *ColorizedLogger, taskId, accountEmail, accessToken string, gid int) {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Get User Home Directory: %s", taskId, err.Error()))
		return
	}

	sessionPath := filepath.Join(home, "Popmart CLI", "sessions.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Read Sessions File: %s", taskId, err.Error()))
		return
	}

	var sessions []Session
	if len(data) > 0 {
		if err := json.Unmarshal(data, &sessions); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Parse Sessions File: %s", taskId, err.Error()))
			return
		}
	}

	found := false
	for i := range sessions {
		if sessions[i].AccountEmail == accountEmail {
			sessions[i].AccessToken = accessToken
			sessions[i].GID = gid
			found = true
			break
		}
	}

	if !found {
		newSession := Session{
			AccountEmail: accountEmail,
			AccessToken:  accessToken,
			GID:          gid,
		}
		sessions = append(sessions, newSession)
	}

	updatedData, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Serialize Session Data: %s", taskId, err.Error()))
		return
	}

	if err := os.WriteFile(sessionPath, updatedData, 0644); err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Save Sessions To Sessions File: %s", taskId, err.Error()))
		return
	}

	logger.Info(fmt.Sprintf("Task %s: Session Successfully Saved For %s", taskId, accountEmail))
}

// ---------------------- WORKER FUNCTION ---------------------- \\
func CalculateWorkers() int {
	numCPU := runtime.NumCPU()
	base := numCPU * 10
	max := int(math.Min(float64(base), 1000))
	return max
}
