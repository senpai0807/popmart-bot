package helpers

type ColorizedLogger struct {
	useColor bool
}

type Task struct {
	TaskId        string
	TaskGroupName string
	Site          string
	Mode          string
	Input         string
	Size          string
	ProfileGroup  string
	AccountGroup  string
	ProxyGroup    string
	Account       string
	Payment       string
	Quantity      int
	Delay         int
	Proxies       []string
	Profile       Profile
}

type Profile struct {
	ProfileName string
	Email       string
	Name        string
	Phone       string
	Address1    string
	Address2    string
	City        string
	PostCode    string
	Country     string
	State       string
	CardNumber  string
	ExpMonth    string
	ExpYear     string
	CVV         string
}

// --------------------- WEBHOOK STRUCT --------------------- \\
type PaypalWebhook struct {
	CheckoutLink string
	Account      string
	Site         string
	Mode         string
	Product      string
	Size         string
	OrderNumber  string
	Profile      string
	ProxyGroup   string
	Image        string
}

type Webhook struct {
	Type        string
	Account     string
	Site        string
	Mode        string
	Product     string
	Size        string
	OrderNumber string
	Profile     string
	ProxyGroup  string
	Image       string
}

// --------------------- SESSIONS STRUCT --------------------- \\
type UserData struct {
	AccessToken string
	GID         int
}

type Session struct {
	AccountEmail string `json:"accountEmail"`
	AccessToken  string `json:"accessToken"`
	GID          int    `json:"gid"`
}
