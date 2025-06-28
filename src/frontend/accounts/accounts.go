package accounts

import (
	"fmt"

	accounts "popmart/src/backend/accounts"
	helpers "popmart/src/middleware/helpers"

	"github.com/AlecAivazis/survey/v2"
)

func AccountsMenu(logger *helpers.ColorizedLogger) {
	for {
		var result string
		options := []string{
			"Add Accounts",
			"Back",
		}

		prompt := &survey.Select{
			Message: "Accounts Menu:",
			Options: options,
		}

		err := survey.AskOne(prompt, &result)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Prompt Accounts Menu: %v", err))
			return
		}

		switch result {
		case "Add Accounts":
			var groupName string
			namePrompt := &survey.Input{
				Message: "Enter Account Group Name:",
			}
			err := survey.AskOne(namePrompt, &groupName)
			if err != nil {
				logger.Error("Prompt Has Failed Or Been Cancelled: " + err.Error())
				continue
			}

			err = accounts.AddAccountGroup(logger, groupName)
			if err != nil {
				logger.Error("Failed To Add Account Group: " + err.Error())
				continue
			}
			logger.Silly("Successfully Saved Account Group ✅")

		case "Back":
			return

		default:
			logger.Warn("Invalid option selected")
		}
	}
}
