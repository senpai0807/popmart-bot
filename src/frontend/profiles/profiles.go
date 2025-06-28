package profiles

import (
	"fmt"

	profiles "popmart/src/backend/profiles"
	helpers "popmart/src/middleware/helpers"

	"github.com/AlecAivazis/survey/v2"
)

func ProfilesMenu(logger *helpers.ColorizedLogger) {
	for {
		var result string
		options := []string{
			"Open Profiles",
			"Back",
		}

		prompt := &survey.Select{
			Message: "Profiles Menu:",
			Options: options,
		}

		err := survey.AskOne(prompt, &result)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Prompt Profiles Menu: %v", err))
			return
		}

		switch result {
		case "Open Profiles":
			err := profiles.OpenProfilesCSV()
			if err != nil {
				logger.Error("Failed To Open Profiles: " + err.Error())
				continue
			}
			logger.Silly("Opened Profiles CSV In Default Editor")

		case "Back":
			return

		default:
			logger.Warn("Invalid option selected")
		}
	}
}
