package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"popmart/src/middleware/helpers"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

func TD(logger *helpers.ColorizedLogger, payload json.RawMessage, taskId, path, proxyUrl, method, userAgent string) (ApiResp, error) {
	logger.Verbose(fmt.Sprintf("Task %s: Generating Trust Decision Parameters", taskId))
	client, err := CreateTLSClient(proxyUrl)
	if err != nil {
		return ApiResp{}, err
	}

	fingerprint, err := RunJS()
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Generate Fingerprint For TD API", taskId))
		return ApiResp{}, err
	}

	form := url.Values{}
	form.Add("data", fingerprint)

	req, err := http.NewRequest("POST", "https://us-fp.apitd.net/web/v2?partner=popmart&appKey=e8e328d27d9866dcf49ed2e0bb7411c4", strings.NewReader(form.Encode()))
	if err != nil {
		return ApiResp{}, err
	}

	req.Header = http.Header{
		"Host":               {"us-fp.apitd.net"},
		"Connection":         {"keep-alive"},
		"sec-ch-ua-platform": {`"Windows"`},
		"User-Agent":         {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"},
		"sec-ch-ua":          {`"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`},
		"Content-Type":       {"application/x-www-form-urlencoded"},
		"sec-ch-ua-mobile":   {"?0"},
		"Accept":             {"*/*"},
		"Origin":             {"https://www.popmart.com"},
		"Sec-Fetch-Site":     {"cross-site"},
		"Sec-Fetch-Mode":     {"cors"},
		"Sec-Fetch-Dest":     {"empty"},
		"Referer":            {"https://www.popmart.com/"},
		"Accept-Encoding":    {"gzip, deflate, br, zstd"},
		"Accept-Language":    {"en-US,en;q=0.9"},
		"Header-Order:": {
			"Host", "Connection", "Content-Length", "sec-ch-ua-platform", "User-Agent", "sec-ch-ua", "Content-Type", "sec-ch-ua-mobile",
			"Accept", "Origin", "Sec-Fetch-Site", "Sec-Fetch-Mode", "Sec-Fetch-Dest", "Referer", "Accept-Encoding", "Accept-Language",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Execute TD API Request", taskId))
		return ApiResp{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Read TD API Response Body", taskId))
		return ApiResp{}, err
	}

	var tdResp TdResp
	if err := json.Unmarshal(respBody, &tdResp); err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal TD API Response Body", taskId))
		return ApiResp{}, err
	}

	if tdResp.Code == "" || tdResp.Result == "" || tdResp.RequestId == "" {
		logger.Error(fmt.Sprintf("Task %s: Missing Required Parameters From TD API", taskId))
		return ApiResp{}, err
	}

	result, err := RunDecode(tdResp)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Decode TD API Response", taskId))
		return ApiResp{}, err
	}

	session, err := RunSession(result.TokenId, path, fmt.Sprintf("%s|https://www.popmart.com/us/collection/1|%s", userAgent, WebGLs[rng.Intn(len(WebGLs))]))
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Parse Session Data From TD API", taskId))
		return ApiResp{}, err
	}

	timestamp := GetTimestamp()
	sign := GenerateSignature(timestamp)

	var bodyMap map[string]any
	if err := json.Unmarshal(payload, &bodyMap); err != nil {
		return ApiResp{}, err
	}
	sData := GenerateSData(bodyMap, timestamp, method, "W_ak^moHpMla")

	s, ok := sData["s"].(string)
	if !ok {
		logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Signature", taskId))
		return ApiResp{}, err
	}

	t, ok := sData["t"].(int)
	if !ok {
		logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Signature", taskId))
		return ApiResp{}, err
	}

	apiResp := ApiResp{
		Success:     true,
		S:           s,
		T:           t,
		Sign:        sign,
		SessionSign: session.SessionSign,
		SessionKey:  session.SessionKey,
	}
	return apiResp, nil
}
