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

func CheckExists(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, accountEmail, proxyUrl string) error {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		orderedData := []OrderedKV{
			{"email", accountEmail},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/customer/v1/customer/exist", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Checking Email Existence", task.TaskId))
		jsonPayload, err := json.Marshal(CheckExistPayload{
			Email: accountEmail,
			S:     tdResp.S,
			T:     int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/customer/v1/customer/exist", strings.NewReader(string(jsonPayload)))
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Create Request, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req.Header = http.Header{
			"language":           {"en"},
			"sec-ch-ua-platform": {`"Windows"`},
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
			"td-session-path":    {"/customer/v1/customer/exist"},
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
				"content-length", "language", "sec-ch-ua-platform", "x-project-id", "x-device-os-type", "sec-ch-ua", "td-session-sign", "sec-ch-ua-mobile", "grey-secret", "accept", "content-type", "td-session-query", "x-client-country", "td-session-key",
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

		var checkResp CheckResp
		if err := json.Unmarshal(respBody, &checkResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch checkResp.Message {
			case "success":
				return nil
			default:
				logger.Error(fmt.Sprintf("Error Checking Account Existence [%s], Retrying [%d]", checkResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Error Checking Account Existence [%d], Retrying [%d]", resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return fmt.Errorf("maxium retries reached")
}

func Login(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, accountEmail, accountPassword, proxyUrl string) (helpers.UserData, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		orderedData := []OrderedKV{
			{"email", accountEmail},
			{"password", accountPassword},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/customer/v1/customer/login", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Logging Into Popmart Account", task.TaskId))
		jsonPayload, err := json.Marshal(LoginPayload{
			Email:    accountEmail,
			Password: accountPassword,
			S:        tdResp.S,
			T:        int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/customer/v1/customer/login", strings.NewReader(string(jsonPayload)))
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Create Request, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req.Header = http.Header{
			"language":           {"en"},
			"sec-ch-ua-platform": {`"Windows"`},
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
			"td-session-path":    {"/customer/v1/customer/login"},
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
				"content-length", "language", "sec-ch-ua-platform", "x-project-id", "x-device-os-type", "sec-ch-ua", "td-session-sign", "sec-ch-ua-mobile", "grey-secret", "accept", "content-type", "td-session-query", "x-client-country", "td-session-key",
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

		var loginResp LoginResp
		if err := json.Unmarshal(respBody, &loginResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch loginResp.Message {
			case "success":
				userData := helpers.UserData{
					AccessToken: loginResp.Data.Token,
					GID:         loginResp.Data.User.Gid,
				}

				helpers.SaveSession(logger, task.TaskId, accountEmail, loginResp.Data.Token, loginResp.Data.User.Gid)
				return userData, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Logging Into Account [%s], Retrying [%d]", task.TaskId, loginResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Logging Into Account [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return helpers.UserData{}, fmt.Errorf("maxium retries reached")
}
