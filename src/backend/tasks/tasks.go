package tasks

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	helpers "popmart/src/middleware/helpers"

	"github.com/google/uuid"
)

func OpenTasksCSV() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve user home directory: %w", err)
	}

	tasksPath := filepath.Join(home, "Popmart CLI", "tasks.csv")

	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		return fmt.Errorf("tasks.csv does not exist at %s", tasksPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", tasksPath)
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "", tasksPath)
	default:
		cmd = exec.Command("xdg-open", tasksPath)
	}

	return cmd.Run()
}

// -------------- LOAD TASKS LOGIC -------------- \\
func LoadTaskGroups() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user home directory: %w", err)
	}

	path := filepath.Join(home, "Popmart CLI", "tasks.csv")

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open tasks.csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV content: %w", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("tasks.csv is empty or missing headers")
	}

	headers := records[0]
	taskGroupIndex := -1
	for i, header := range headers {
		if header == "Task Group" {
			taskGroupIndex = i
			break
		}
	}
	if taskGroupIndex == -1 {
		return nil, fmt.Errorf(`"Task Group" column not found in tasks.csv`)
	}

	groupMap := map[string]bool{}
	for _, row := range records[1:] {
		if len(row) > taskGroupIndex {
			group := row[taskGroupIndex]
			if group != "" {
				groupMap[group] = true
			}
		}
	}

	var uniqueGroups []string
	for group := range groupMap {
		uniqueGroups = append(uniqueGroups, group)
	}

	return uniqueGroups, nil
}

func LoadTasks(logger *helpers.ColorizedLogger, groupName string) ([]helpers.Task, error) {
	if err := LoadJson(ProxiesPath, &ProxyGroups); err != nil {
		return nil, err
	}

	if err := LoadJson(AccountsPath, &AccountGroups); err != nil {
		return nil, err
	}

	profileRecords, err := LoadCsv(ProfilesPath)
	if err != nil {
		return nil, err
	}

	profileGroups, err := BuildProfile(profileRecords)
	if err != nil {
		return nil, err
	}

	taskRecords, err := LoadCsv(TasksPath)
	if err != nil {
		return nil, err
	}

	if len(taskRecords) < 2 {
		return nil, fmt.Errorf("tasks.csv is empty or missing headers")
	}

	headers := taskRecords[0]
	indexMap := make(map[string]int)
	for i, h := range headers {
		indexMap[h] = i
	}

	required := []string{
		"Task Group", "Site", "Mode", "Input", "Size",
		"Proxy Group", "Profile Group", "Account Group", "Quantity",
		"Delay", "Payment Method",
	}
	for _, col := range required {
		if _, ok := indexMap[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var tasks []helpers.Task

	for _, row := range taskRecords[1:] {
		if len(row) < len(headers) || row[indexMap["Task Group"]] != groupName {
			continue
		}

		profileGroup := row[indexMap["Profile Group"]]
		accountGroup := row[indexMap["Account Group"]]
		proxyGroup := row[indexMap["Proxy Group"]]

		delay := ParseInt(row[indexMap["Delay"]], 3500)
		quantity := ParseInt(row[indexMap["Quantity"]], 1)
		size := Normalize(row[indexMap["Size"]])

		pg := FindProxy(ProxyGroups, proxyGroup)
		if pg == nil {
			logger.Error(fmt.Sprintf("Proxy Group Not Found: %s", proxyGroup))
			continue
		}

		profiles := profileGroups[profileGroup]
		if len(profiles) == 0 {
			logger.Error(fmt.Sprintf("Profile Group Not Found Or Empty: %s", profileGroup))
			continue
		}

		for _, profile := range profiles {
			account := FindAccount(AccountGroups, accountGroup, profile.Email)
			if account == "" {
				logger.Warn(fmt.Sprintf("No Matching Account For Email: %s", profile.Email))
				continue
			}

			tasks = append(tasks, helpers.Task{
				TaskId:        uuid.New().String(),
				TaskGroupName: row[indexMap["Task Group"]],
				Site:          row[indexMap["Site"]],
				Mode:          row[indexMap["Mode"]],
				Input:         row[indexMap["Input"]],
				Size:          size,
				ProfileGroup:  profileGroup,
				AccountGroup:  accountGroup,
				ProxyGroup:    pg.Name,
				Proxies:       pg.Proxies,
				Account:       account,
				Payment:       row[indexMap["Payment Method"]],
				Quantity:      quantity,
				Delay:         delay,
				Profile:       profile,
			})
		}
	}

	logger.Silly(fmt.Sprintf("Loaded %d Tasks From Group %s", len(tasks), groupName))
	return tasks, nil
}
