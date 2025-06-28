package desktop

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	helpers "popmart/src/middleware/helpers"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

func FetchCheckoutId(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, proxyUrl string) (string, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		logger.Verbose(fmt.Sprintf("Task %s: Fetching Adyen Checkout ID", task.TaskId))
		jsonPayload, err := json.Marshal(map[string]any{
			"experiments": []string{},
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://checkoutshopper-live.adyen.com/checkoutshopper/v2/analytics/id?clientKey=live_T4D4ECRSB5G3DHDXMJHYRUDRP4ER4U52", strings.NewReader(string(jsonPayload)))
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Create Request, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req.Header = http.Header{
			"accept":             {"application/json, text/plain, */*"},
			"accept-encoding":    {"gzip, deflate, br, zstd"},
			"accept-language":    {"en-US,en;q=0.9"},
			"connection":         {"keep-alive"},
			"content-type":       {"application/json"},
			"host":               {"checkoutshopper-live.adyen.com"},
			"origin":             {"https://popmart.com"},
			"referer":            {"https://www.popmart.com/us/checkout?type=normal"},
			"sec-ch-ua":          {helpers.SecChUa},
			"sec-ch-ua-mobile":   {"?0"},
			"sec-ch-ua-platform": {`"Windows"`},
			"sec-fetch-dest":     {"empty"},
			"sec-fetch-mode":     {"cors"},
			"sec-fetch-site":     {"cross-site"},
			"user-agent":         {helpers.UserAgent},
			"Header-Order:": {
				"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "upgrade-insecure-requests", "user-agent", "accept",
				"sec-fetch-site", "sec-fetch-mode", "sec-fetch-user", "sec-fetch-dest", "accept-encoding", "accept-language", "priority",
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

		var adyenResponse AdyenResponse
		if err := json.Unmarshal(respBody, &adyenResponse); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			return adyenResponse.ID, nil
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Fetching Adyen Checkout ID [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return "", fmt.Errorf("maxium retries reached")
}
