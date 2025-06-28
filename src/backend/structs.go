package backend

type AccountGroup struct {
	Name     string   `json:"name"`
	ID       string   `json:"id"`
	Accounts []string `json:"accounts"`
}

type ProxyGroup struct {
	Name    string   `json:"name"`
	ID      string   `json:"id"`
	Proxies []string `json:"proxies"`
}

type Settings struct {
	ImapEmail    string `json:"imapEmail"`
	ImapPassword string `json:"imapPassword"`
	WebhookUrl   string `json:"webhookUrl"`
}
