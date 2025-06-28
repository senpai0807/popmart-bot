package settings

import (
	"fmt"

	backend "popmart/src/backend"
	setting "popmart/src/backend/settings"
	helpers "popmart/src/middleware/helpers"

	"github.com/AlecAivazis/survey/v2"
)

func SettingsMenu(logger *helpers.ColorizedLogger) {
	for {
		var result string
		options := []string{
			"Add IMAP",
			"Add Webhook",
			"Test Webhook",
			"Back",
		}

		prompt := &survey.Select{
			Message: "Settings Menu:",
			Options: options,
		}

		err := survey.AskOne(prompt, &result)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Prompt Settings Menu: %v", err))
			return
		}

		switch result {
		case "Add IMAP":
			var email, password string
			emailPrompt := &survey.Input{
				Message: "Enter IMAP Email:",
			}

			if err := survey.AskOne(emailPrompt, &email); err != nil {
				logger.Error("Prompt Has Failed Or Been Cancelled: " + err.Error())
				continue
			}

			passwordPrompt := &survey.Password{
				Message: "Enter IMAP Password:",
			}

			if err := survey.AskOne(passwordPrompt, &password); err != nil {
				logger.Error("Prompt Has Failed Or Been Cancelled: " + err.Error())
				continue
			}

			if err := setting.AddImap(logger, email, password); err != nil {
				logger.Error("Failed To Save IMAP Settings: " + err.Error())
				continue
			}
			logger.Silly("Successfully Saved IMAP Settings")
		case "Add Webhook":
			var url string
			input := &survey.Input{
				Message: "Enter Discord Webhook URL:",
			}
			if err := survey.AskOne(input, &url); err != nil {
				logger.Error("Prompt Has Failed Or Been Cancelled: " + err.Error())
				continue
			}

			if err := setting.UpdateWebhookURL(logger, url); err != nil {
				logger.Error("Failed To Update Discord Webhook: " + err.Error())
				continue
			}
			logger.Silly("Successfully Saved Discord Webhook Settings")

		case "Test Webhook":
			settings, err := backend.LoadSettings()
			if err != nil {
				logger.Error("Failed To Load Settings: " + err.Error())
				return
			}

			if settings.WebhookUrl == "" {
				logger.Warn("No Webhook URL Was Found In Settings File")
				return
			}

			setting.SendTestWebhook(settings)
			logger.Silly("Webhook Test Successfully Sent ✅")

		case "Back":
			return

		default:
			logger.Warn("Invalid option selected")
		}
	}
}
