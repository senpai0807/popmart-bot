package discord

import (
	"encoding/json"
	"fmt"
	"time"

	backend "popmart/src/backend"
	helpers "popmart/src/middleware/helpers"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
)

func SendPaypal(logger *helpers.ColorizedLogger, data helpers.PaypalWebhook, taskId string) error {
	settings, err := backend.LoadSettings()
	if err != nil {
		return err
	}

	hook := discordwebhook.Hook{
		Username:   "Popmart CLI",
		Avatar_url: "https://i.imgur.com/JWAP07j.jpeg",
		Embeds: []discordwebhook.Embed{
			{
				Title:     "Popmart Checkout Link Ready",
				Url:       data.CheckoutLink,
				Color:     5662170,
				Timestamp: time.Now(),
				Thumbnail: discordwebhook.Thumbnail{Url: data.Image},
				Fields: []discordwebhook.Field{
					{Name: "**Account**", Value: data.Account, Inline: false},
					{Name: "**Site**", Value: data.Site, Inline: true},
					{Name: "**Mode**", Value: data.Mode, Inline: true},
					{Name: "**Product**", Value: data.Product, Inline: true},
					{Name: "**Size**", Value: data.Size, Inline: true},
					{Name: "**Profile**", Value: data.Profile, Inline: true},
					{Name: "**Proxy Group**", Value: data.ProxyGroup, Inline: true},
					{Name: "**Order Number**", Value: data.OrderNumber, Inline: false},
				},
				Footer: discordwebhook.Footer{
					Text:     "Popmart CLI",
					Icon_url: "https://i.imgur.com/JWAP07j.jpeg",
				},
			},
		},
	}

	if err := helpers.PlaySound(helpers.Success); err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Play Checkout Sound: %v", taskId, err))
	}

	payload, err := json.Marshal(hook)
	if err != nil {
		return err
	}
	return discordwebhook.ExecuteWebhook(settings.WebhookUrl, payload)
}

func SendWebhook(logger *helpers.ColorizedLogger, data helpers.Webhook, taskId string) error {
	var (
		title string
		color int
	)
	settings, err := backend.LoadSettings()
	if err != nil {
		return err
	}

	switch data.Type {
	case "Success":
		title = "Checkout Success 🌙"
		color = 5662170
		if err := helpers.PlaySound(helpers.Success); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Play Checkout Sound: %v", taskId, err))
		}
	case "Failure":
		title = "Checkout Failed ❌"
		color = 8388640
		if err := helpers.PlaySound(helpers.Decline); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Play Decline Sound: %v", taskId, err))
		}
	default:
		return fmt.Errorf("unsupported type")
	}

	hook := discordwebhook.Hook{
		Username:   "Popmart CLI",
		Avatar_url: "https://i.imgur.com/JWAP07j.jpeg",
		Embeds: []discordwebhook.Embed{
			{
				Title:     title,
				Color:     color,
				Timestamp: time.Now(),
				Thumbnail: discordwebhook.Thumbnail{Url: data.Image},
				Fields: []discordwebhook.Field{
					{Name: "**Account**", Value: data.Account, Inline: false},
					{Name: "**Site**", Value: data.Site, Inline: true},
					{Name: "**Mode**", Value: data.Mode, Inline: true},
					{Name: "**Product**", Value: data.Product, Inline: true},
					{Name: "**Size**", Value: data.Size, Inline: true},
					{Name: "**Profile**", Value: data.Profile, Inline: true},
					{Name: "**Proxy Group**", Value: data.ProxyGroup, Inline: true},
					{Name: "**Order Number**", Value: data.OrderNumber, Inline: false},
				},
				Footer: discordwebhook.Footer{
					Text:     "Popmart CLI",
					Icon_url: "https://i.imgur.com/JWAP07j.jpeg",
				},
			},
		},
	}

	payload, err := json.Marshal(hook)
	if err != nil {
		return err
	}
	return discordwebhook.ExecuteWebhook(settings.WebhookUrl, payload)
}
