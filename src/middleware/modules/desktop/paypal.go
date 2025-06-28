package desktop

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	helpers "popmart/src/middleware/helpers"
	api "popmart/src/middleware/helpers/api"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

func Paypal(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, order OrderDetails, accountEmail, proxyUrl string) (helpers.PaypalWebhook, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		ordered := OrderedMap{
			{"orderNo", order.OrderNumber},
			{"saveCard", false},
			{"returnURL", "https://www.popmart.com/us/checkout"},
			{"cancelURL", "https://www.popmart.com/us/checkout"},
		}

		data, err := ordered.MarshalJSON()
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/shop/v1/shop/cash/desk/paypal/pay", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Warn(fmt.Sprintf("Task %s: Creating Paypal Checkout Link", task.TaskId))
		jsonPayload, err := json.Marshal(PaypalPayload{
			OrderNo:   order.OrderNumber,
			SaveCard:  false,
			ReturnUrl: "https://www.popmart.com/us/checkout",
			CancelUrl: "https://www.popmart.com/us/checkout",
			S:         tdResp.S,
			T:         int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/shop/v1/shop/cash/desk/paypal/pay", strings.NewReader(string(jsonPayload)))
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Create Request, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req.Header = http.Header{
			"language":           {"en"},
			"sec-ch-ua-platform": {`"Windows"`},
			"authorization":      {fmt.Sprintf("Bearer %s", userData.AccessToken)},
			"x-project-id":       {"naus"},
			"x-device-os-type":   {"web"},
			"sec-ch-ua":          {helpers.SecChUa},
			"td-session-sign":    {tdResp.SessionSign},
			"sec-ch-ua-mobile":   {"?0"},
			"grey-secret":        {"null"},
			"accept":             {"application/json, text/plain, */*"},
			"content-type":       {"application/json"},
			"td-session-query":   {""},
			"x-client-country":   {"US"},
			"td-session-key":     {tdResp.SessionKey},
			"tz":                 {"America/New_York"},
			"td-session-path":    {"/shop/v1/shop/cash/desk/paypal/pay"},
			"country":            {"US"},
			"x-sign":             {tdResp.Sign},
			"clientkey":          {"nw3b089qrgw9m7b7i"},
			"user-agent":         {helpers.UserAgent},
			"x-client-namespace": {"america"},
			"origin":             {"https://www.popmart.com"},
			"sec-fetch-site":     {"same-site"},
			"sec-fetch-mode":     {"cors"},
			"sec-fetch-dest":     {"empty"},
			"referer":            {"https://www.popmart.com/"},
			"accept-encoding":    {"gzip, deflate, br, zstd"},
			"accept-language":    {"en-US,en;q=0.9"},
			"priority":           {"u=1, i"},
			"Header-Order:": {
				"content-length", "language", "sec-ch-ua-platform", "authorization", "x-project-id", "x-device-os-type", "sec-ch-ua", "td-session-sign", "sec-ch-ua-mobile", "grey-secret", "accept", "content-type", "td-session-query", "x-client-country", "td-session-key",
				"tz", "td-session-path", "country", "x-sign", "clientkey", "user-agent", "x-client-namespace", "origin", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language", "priority",
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Execute Request, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Read Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		var paypalResp PaypalResp
		if err := json.Unmarshal(respBody, &paypalResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch paypalResp.Message {
			case "success":
				paypal := helpers.PaypalWebhook{
					CheckoutLink: fmt.Sprintf("https://www.paypal.com/checkoutnow?token=%s&fundingSource=paypal&redirect_uri=sdk.ios.paypal://x-callback-url/paypal-sdk/paypal-checkout&native_xo=1", paypalResp.Data.PlatformOrderNum),
					Account:      accountEmail,
					Site:         task.Site,
					Mode:         task.Mode,
					Product:      order.ProductName,
					Size:         task.Size,
					OrderNumber:  order.OrderNumber,
					Profile:      task.Profile.ProfileName,
					ProxyGroup:   task.ProxyGroup,
					Image:        order.ProductImage,
				}

				helpers.IncrementCheckedOut()
				logger.Silly(fmt.Sprintf("Task %s: Successful Checkout 🌙", task.TaskId))
				return paypal, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Creating Paypal Checkout Link [%s], Retrying [%d]", task.TaskId, paypalResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Creating Paypal Checkout Link [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return helpers.PaypalWebhook{}, fmt.Errorf("maxium retries reached")
}
