package accounts

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	backend "popmart/src/backend"
	helpers "popmart/src/middleware/helpers"

	"github.com/google/uuid"
)

func AddAccountGroup(logger *helpers.ColorizedLogger, groupName string) error {
	tmpFile, err := os.CreateTemp("", "accounts_*.txt")
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

	var accounts []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			logger.Warn(fmt.Sprintf("Line %d Skipped: Missing Colon Separator", lineNum))
			continue
		}

		email := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])

		if email == "" || password == "" {
			logger.Warn(fmt.Sprintf("Line %d Skipped: Empty Email Or Password", lineNum))
			continue
		}

		if _, err := mail.ParseAddress(email); err != nil {
			logger.Warn(fmt.Sprintf("Line %d Skipped: Invalid Email Format: %s", lineNum, email))
			continue
		}

		accounts = append(accounts, fmt.Sprintf("%s:%s", email, password))
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

	path := filepath.Join(home, "Popmart CLI", "accounts.json")

	var accountGroups []backend.AccountGroup

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &accountGroups)
	}

	accountGroups = append(accountGroups, backend.AccountGroup{
		Name:     groupName,
		ID:       uuid.New().String(),
		Accounts: accounts,
	})

	updated, err := json.MarshalIndent(accountGroups, "", "  ")
	if err != nil {
		logger.Error("Failed To Marshal Accounts JSON")
		return err
	}

	err = os.WriteFile(path, updated, 0644)
	if err != nil {
		logger.Error("Failed To Write To Accounts File")
		return err
	}

	return nil
}
