package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	backend "popmart/src/backend"
	helpers "popmart/src/middleware/helpers"
)

func UpdateWebhookURL(logger *helpers.ColorizedLogger, webhook string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed To Get Home Directory")
		return err
	}

	settingsPath := filepath.Join(home, "Popmart CLI", "settings.json")

	var settings map[string]string
	fileData, err := os.ReadFile(settingsPath)
	if err != nil {
		logger.Error("Failed To Read Settings File")
		return err
	}

	err = json.Unmarshal(fileData, &settings)
	if err != nil {
		logger.Error("Failed To Unmarshal Settings File")
		return err
	}

	settings["webhookUrl"] = webhook
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(settingsPath, data, 0644)
	if err != nil {
		logger.Error("Failed To Update Settings File")
		return err
	}

	return nil
}

func SendTestWebhook(settings backend.Settings) {
	webhookData := func(title, color string) map[string]any {
		return map[string]any{
			"username":   "Popmart CLI",
			"avatar_url": "https://i.imgur.com/JWAP07j.jpeg",
			"embeds": []map[string]any{
				{
					"title": title,
					"color": backend.ParseIntColor(color),
					"footer": map[string]string{
						"text":     "Popmart CLI",
						"icon_url": "https://i.imgur.com/JWAP07j.jpeg",
					},
					"timestamp": time.Now().Format(time.RFC3339),
				},
			},
		}
	}

	webhooks := []struct {
		url   string
		title string
		color string
	}{
		{settings.WebhookUrl, "Webhook Test 🌙", "#5665DA"},
	}

	for _, webhook := range webhooks {
		if webhook.url != "" {
			go backend.SendWebhook(webhook.url, webhookData(webhook.title, webhook.color))
		}
	}
}

func AddImap(logger *helpers.ColorizedLogger, email, password string) error {
	normalizedPassword := ""
	for _, c := range password {
		if c != ' ' {
			normalizedPassword += string(c)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed To Get Home Directory")
		return err
	}
	settingsPath := filepath.Join(home, "Popmart CLI", "settings.json")

	var settings map[string]string
	fileData, err := os.ReadFile(settingsPath)
	if err != nil {
		logger.Error("Failed To Read Settings File")
		return err
	}

	err = json.Unmarshal(fileData, &settings)
	if err != nil {
		logger.Error("Failed To Unmarshal Settings File")
		return err
	}

	settings["imapEmail"] = email
	settings["imapPassword"] = normalizedPassword

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		logger.Error("Failed To Marshal Updated Settings")
		return err
	}

	err = os.WriteFile(settingsPath, data, 0644)
	if err != nil {
		logger.Error("Failed To Update Settings File")
		return err
	}

	return nil
}
