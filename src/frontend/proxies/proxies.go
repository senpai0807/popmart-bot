package proxies

import (
	"fmt"

	backend "popmart/src/backend"
	proxies "popmart/src/backend/proxies"
	helpers "popmart/src/middleware/helpers"

	"github.com/AlecAivazis/survey/v2"
)

func ProxiesMenu(logger *helpers.ColorizedLogger) {
	for {
		var result string
		options := []string{
			"Add Proxies",
			"Test Proxies",
			"Back",
		}

		prompt := &survey.Select{
			Message: "Proxies Menu:",
			Options: options,
		}

		err := survey.AskOne(prompt, &result)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Prompt Proxies Menu: %v", err))
			return
		}

		switch result {
		case "Add Proxies":
			var groupName string
			namePrompt := &survey.Input{
				Message: "Enter Proxy Group Name:",
			}
			if err := survey.AskOne(namePrompt, &groupName); err != nil {
				logger.Error("Prompt Has Failed Or Been Cancelled: " + err.Error())
				continue
			}

			err = proxies.AddProxyGroup(logger, groupName)
			if err != nil {
				logger.Error("Failed To Add Proxy Group: " + err.Error())
				continue
			}

			logger.Silly("Successfully Saved Proxy Group ✅")

		case "Test Proxies":
			groups, err := backend.LoadProxyGroups()
			if err != nil {
				logger.Error("Failed To Load Proxy Groups: " + err.Error())
				continue
			}

			if len(groups) == 0 {
				logger.Warn("No Proxy Groups Found")
				continue
			}

			var groupName string
			var names []string
			for _, g := range groups {
				names = append(names, g.Name)
			}

			groupPrompt := &survey.Select{
				Message: "Select Proxy Group To Test:",
				Options: names,
			}
			if err := survey.AskOne(groupPrompt, &groupName); err != nil {
				logger.Error("Prompt Cancelled Or Failed: " + err.Error())
				continue
			}

			for _, group := range groups {
				if group.Name == groupName {
					proxies.TestProxyGroup(logger, group)
					break
				}
			}

		case "Back":
			return

		default:
			logger.Warn("Invalid option selected")
		}
	}
}
