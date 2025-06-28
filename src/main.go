package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	accounts "popmart/src/frontend/accounts"
	profiles "popmart/src/frontend/profiles"
	proxies "popmart/src/frontend/proxies"
	settings "popmart/src/frontend/settings"
	tasks "popmart/src/frontend/tasks"
	helpers "popmart/src/middleware/helpers"
	update "popmart/src/middleware/helpers/update"

	"github.com/AlecAivazis/survey/v2"
)

func main() {
	logger := helpers.NewColorizedLogger(true)

	var ver update.VersionInfo
	if err := json.Unmarshal(update.VersionData, &ver); err != nil {
		os.Exit(1)
	}

	logger.Info("Checking For Available Updates")
	updateAvailable, err := update.CheckAndUpdate(ver.Version)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed To Check For Available Updates: %v", err))
		return
	}

	if updateAvailable {
		logger.Info("Downloading Latest Update")

		err := update.Update()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Update To Latest Version: %v", err))
			return
		}
	}

	title := fmt.Sprintf("v%s | Carted: 0 | Secured: 0", ver.Version)
	fmt.Printf("\033]0;%s\007", title)

	helpers.InitFileSystem(logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Exiting Popmart CLI 👋")
		os.Exit(0)
	}()

	logger.Info("You're On The Latest Version, Welcome, User!")

	for {
		options := []string{
			"Tasks",
			"Proxies",
			"Profiles",
			"Accounts",
			"Settings",
			"Exit",
		}

		var result string
		prompt := &survey.Select{
			Message: "Select an Option:",
			Options: options,
		}

		err := survey.AskOne(prompt, &result)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Prompt CLI Menu: %v", err))
			return
		}

		switch result {
		case "Tasks":
			tasks.TasksMenu(logger)
		case "Proxies":
			proxies.ProxiesMenu(logger)
		case "Profiles":
			profiles.ProfilesMenu(logger)
		case "Accounts":
			accounts.AccountsMenu(logger)
		case "Settings":
			settings.SettingsMenu(logger)
		case "Exit":
			fmt.Println("Exiting Popmart CLI 👋")
			return
		default:
			logger.Warn("Unknown selection.")
		}
	}
}
