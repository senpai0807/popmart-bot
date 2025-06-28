package desktop

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	helpers "popmart/src/middleware/helpers"
	api "popmart/src/middleware/helpers/api"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

func FetchProduct(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, proxyUrl string) (ProductDetails, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		orderedData := []OrderedKV{
			{"spuId", task.Input},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/shop/v1/shop/productDetails", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Fetching Product Information [%s]", task.TaskId, task.Input))
		req, err := http.NewRequest("GET", fmt.Sprintf("https://prod-na-api.popmart.com/shop/v1/shop/productDetails?spuId=%s&s=%s&t=%d", task.Input, tdResp.S, tdResp.T), nil)
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
			"td-session-path":    {"/shop/v1/shop/productDetails"},
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

		var productResp ProductResp
		if err := json.Unmarshal(respBody, &productResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch productResp.Message {
			case "success":
				product := productResp.Data
				for _, sku := range product.Skus {
					if strings.EqualFold(sku.Title, task.Size) {
						if sku.Stock.OnlineStock == 0 {
							logger.Error(fmt.Sprintf("Task %s: Fetching Product Details [OOS], Retrying [%d]", task.TaskId, retryCount+1))
							helpers.Delay(task.Delay)
							retryCount++
							continue
						}

						productDetails := ProductDetails{
							ProductName: product.Title,
							SpuId:       product.ID,
							SkuId:       sku.ID,
							SkuTitle:    sku.Title,
							MainImage:   sku.MainImage,
							Price:       sku.Price,
							Quantity:    task.Quantity,
						}
						return productDetails, nil
					}
				}

				logger.Error(fmt.Sprintf("Task %s: Error Fetching Product Information [No Matching Size], Retrying [%d]", task.TaskId, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Fetching Product Information [%s], Retrying [%d]", task.TaskId, productResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Fetching Product Information [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return ProductDetails{}, fmt.Errorf("maxium retries reached")
}

func AddToCart(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, productDetails ProductDetails, proxyUrl string) error {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		spuId, err := strconv.ParseInt(productDetails.SpuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		skuId, err := strconv.ParseInt(productDetails.SkuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		orderedData := []OrderedKV{
			{"skuId", skuId},
			{"spuId", spuId},
			{"offsetCount", productDetails.Quantity},
			{"GID", fmt.Sprintf("%d", userData.GID)},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "shop/v1/shoppingcart/offsetAdjustShoppingCartSKUNum", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Adding To Cart", task.TaskId))
		jsonPayload, err := json.Marshal(AtcPayload{
			SkuId:       int(skuId),
			SpuId:       int(spuId),
			OffsetCount: productDetails.Quantity,
			GID:         fmt.Sprintf("%d", userData.GID),
			S:           tdResp.S,
			T:           int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/shop/v1/shoppingcart/offsetAdjustShoppingCartSKUNum", strings.NewReader(string(jsonPayload)))
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
			"td-session-path":    {"shop/v1/shoppingcart/offsetAdjustShoppingCartSKUNum"},
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

		var atcResp AtcResp
		if err := json.Unmarshal(respBody, &atcResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch atcResp.Message {
			case "success":
				helpers.IncrementCarted()
				return nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Adding To Cart [%s], Retrying [%d]", task.TaskId, atcResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Adding To Cart [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return fmt.Errorf("maxium retries reached")
}

func FetchAddress(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, proxyUrl string) (CustomerAddress, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		orderedData := []OrderedKV{}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/customer/v1/address/list", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Checking For Default Address", task.TaskId))
		req, err := http.NewRequest("GET", fmt.Sprintf("https://prod-na-api.popmart.com/customer/v1/address/list?s=%s&t=%d", tdResp.S, tdResp.T), nil)
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
			"td-session-path":    {"/customer/v1/address/list"},
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

		var addressResp DefaultResp
		if err := json.Unmarshal(respBody, &addressResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch addressResp.Message {
			case "success":
				if len(addressResp.Data.List) == 0 {
					return CustomerAddress{}, fmt.Errorf("no default address found")
				}

				var customerAddress CustomerAddress
				found := false
				for _, addr := range addressResp.Data.List {
					if addr.IsDefault {
						customerAddress = CustomerAddress{
							AddressId: addr.ID,
							UserId:    addr.UserID,
							State:     addr.ProvinceName,
							Line1:     addr.DetailInfo,
							Line2:     addr.ExtraAddress,
							City:      addr.CityName,
							PostCode:  addr.PostalCode,
							Phone:     addr.TelNumber,
							FirstName: addr.GivenName,
							LastName:  addr.FamilyName,
						}

						found = true
						break
					}
				}

				if !found {
					return CustomerAddress{}, fmt.Errorf("no default address found")
				}
				return customerAddress, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Fetching Default Address [%s], Retrying [%d]", task.TaskId, addressResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Fetching Default Address [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return CustomerAddress{}, fmt.Errorf("maxium retries reached")
}

func AddAddress(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, proxyUrl string) (CustomerAddress, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		sNameParts := strings.SplitN(task.Profile.Name, " ", 2)
		sFirst, sLast := sNameParts[0], ""
		if len(sNameParts) > 1 {
			sLast = sNameParts[1]
		}

		removeNonDigits := func(input string) string {
			re := regexp.MustCompile(`\D`)
			return re.ReplaceAllString(input, "")
		}

		orderedData := []OrderedKV{
			{"address", []OrderedKV{
				{"givenName", sFirst},
				{"familyName", sLast},
				{"telNumber", removeNonDigits(task.Profile.Phone)},
				{"detailInfo", task.Profile.Address1},
				{"extraAddress", task.Profile.Address2},
				{"cityName", task.Profile.City},
				{"postalCode", task.Profile.PostCode},
				{"isDefault", true},
				{"nationalCode", "US"},
				{"userName", task.Profile.Name},
				{"countryName", "United States"},
				{"provinceName", task.Profile.State},
				{"provinceCode", helpers.StateAbbreviations[task.Profile.State]},
			}},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/customer/v1/address/add", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Submitting Address Information", task.TaskId))
		jsonPayload, err := json.Marshal(AddressPayload{
			Address: AddressData{
				FirstName:    sFirst,
				LastName:     sLast,
				Phone:        removeNonDigits(task.Profile.Phone),
				Line1:        task.Profile.Address1,
				Line2:        task.Profile.Address2,
				City:         task.Profile.City,
				PostalCode:   task.Profile.PostCode,
				IsDefault:    true,
				NationalCode: "US",
				FullName:     task.Profile.Name,
				Country:      "United States",
				ProvinceName: task.Profile.State,
				ProvinceCode: helpers.StateAbbreviations[task.Profile.State],
			},
			S: tdResp.S,
			T: int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/customer/v1/address/add", strings.NewReader(string(jsonPayload)))
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
			"td-session-path":    {"/customer/v1/address/add"},
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

		var addressResp AddressResp
		if err := json.Unmarshal(respBody, &addressResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch addressResp.Message {
			case "success":
				customerAddress := CustomerAddress{
					AddressId: addressResp.Data.Address.ID,
					UserId:    addressResp.Data.Address.UserID,
					State:     task.Profile.State,
					Line1:     task.Profile.Address1,
					Line2:     task.Profile.Address2,
					City:      task.Profile.City,
					PostCode:  task.Profile.PostCode,
					Phone:     removeNonDigits(task.Profile.Phone),
					FirstName: sFirst,
					LastName:  sLast,
				}
				return customerAddress, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Submitting Address Information [%s], Retrying [%d]", task.TaskId, addressResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Submitting Address Information [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return CustomerAddress{}, fmt.Errorf("maxium retries reached")
}

func FetchRates(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, productDetails ProductDetails, proxyUrl string) (int, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		spuId, err := strconv.ParseInt(productDetails.SpuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		skuId, err := strconv.ParseInt(productDetails.SkuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		placeOrderReq := []OrderedKV{
			{"userId", userData.GID},
			{"paymentChannel", -1},
			{"skuItem", []any{
				map[string]any{
					"spuId":               spuId,
					"skuId":               skuId,
					"count":               productDetails.Quantity,
					"skuCount":            productDetails.Quantity,
					"price":               productDetails.Price,
					"title":               task.Size,
					"spuTitle":            productDetails.ProductName,
					"discountPrice":       productDetails.Price,
					"currentSKUInCartNum": productDetails.Quantity,
				},
			}},
			{"mpUserCouponID", nil},
			{"userCouponID", nil},
			{"DiscountCode", nil},
			{"orderTotalAmount", -1},
			{"totalAmount", productDetails.Price},
			{"currency", "USD"},
		}

		orderedData := []OrderedKV{
			{"placeOrderReq", placeOrderReq},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/shop/v1/freight/result", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Fetching Shipping Rates", task.TaskId))
		jsonPayload, err := json.Marshal(RatePayload{
			PlaceOrderReq: PlaceOrderReq{
				UserId:         userData.GID,
				PaymentChannel: -1,
				SkuItem: []RateItem{
					{
						SpuId:           spuId,
						SkuId:           skuId,
						Count:           productDetails.Quantity,
						SkuCount:        productDetails.Quantity,
						Price:           productDetails.Price,
						Title:           task.Size,
						SpuTitle:        productDetails.ProductName,
						DiscountedPrice: productDetails.Price,
						Cart:            productDetails.Quantity,
					},
				},
				MpUserCouponId:   nil,
				UserCouponId:     nil,
				DiscountCode:     nil,
				OrderTotalAmount: -1,
				TotalAmount:      productDetails.Price,
				Currency:         "USD",
			},
			S: tdResp.S,
			T: int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/shop/v1/freight/result", strings.NewReader(string(jsonPayload)))
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
			"td-session-path":    {"/shop/v1/freight/result"},
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

		var rateResp RateResp
		if err := json.Unmarshal(respBody, &rateResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch rateResp.Message {
			case "success":
				if rateResp.Data.DiscountList == nil {
					if len(rateResp.Data.ExpressList) > 0 {
						expressPrice := rateResp.Data.ExpressList[0].ExpressPrice
						return expressPrice, nil
					} else {
						logger.Error(fmt.Sprintf("Task %s: Express List Was Empty In Rate Response", task.TaskId))
						return 0, fmt.Errorf("express list empty")
					}
				} else {
					if discount, ok := rateResp.Data.DiscountList["STANDARD"]; ok {
						if len(discount.List) > 0 {
							freePrice := discount.List[0].DiscountAmount
							return freePrice, nil
						} else {
							logger.Error(fmt.Sprintf("Task %s: Discount List Was Empty", task.TaskId))
							return 0, fmt.Errorf("discount list empty")
						}
					} else {
						logger.Error(fmt.Sprintf("Task %s: STANDARD Discount Not Found", task.TaskId))
						return 0, fmt.Errorf("discount list missing STANDARD key")
					}
				}
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Fetching Shipping Rates [%s], Retrying [%d]", task.TaskId, rateResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Fetching Shipping Rates [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return 0, fmt.Errorf("maxium retries reached")
}

func CalculateTaxes(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, product ProductDetails, customer CustomerAddress, proxyUrl string) (int, int, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		spuId, err := strconv.ParseInt(product.SpuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		skuId, err := strconv.ParseInt(product.SkuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		orderedData := []OrderedKV{
			{"userId", customer.UserId},
			{"AddressId", customer.AddressId},
			{"skuItem", []any{
				map[string]any{
					"spuId":               spuId,
					"skuId":               skuId,
					"count":               product.Quantity,
					"skuCount":            product.Quantity,
					"price":               product.Price,
					"title":               task.Size,
					"spuTitle":            product.ProductName,
					"discountPrice":       product.Price,
					"currentSKUInCartNum": product.Quantity,
				},
			}},
			{"activities", []any{}},
			{"currency", "USD"},
		}

		data, err := MarshalOrderedMap(orderedData)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/shop/v1/shop/calculateOrderAmountMix", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Calculating Taxes", task.TaskId))
		jsonPayload, err := json.Marshal(TaxesPayload{
			UserId:    customer.UserId,
			AddressId: customer.AddressId,
			SkuItem: []TaxesItem{
				{
					SpuId:           spuId,
					SkuId:           skuId,
					Count:           product.Quantity,
					SkuCount:        product.Quantity,
					Price:           product.Price,
					Title:           task.Size,
					SpuTitle:        product.ProductName,
					DiscountedPrice: product.Price,
					Cart:            product.Quantity,
				},
			},
			Activities: []any{},
			Currency:   "USD",
			S:          tdResp.S,
			T:          int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/shop/v1/shop/calculateOrderAmountMix", strings.NewReader(string(jsonPayload)))
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
			"td-session-path":    {"/shop/v1/shop/calculateOrderAmountMix"},
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

		var taxResp TaxResp
		if err := json.Unmarshal(respBody, &taxResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch taxResp.Message {
			case "success":
				return taxResp.Data.TaxAmount, taxResp.Data.TotalAmount, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Calculating Taxes [%s], Retrying [%d]", task.TaskId, taxResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Calculating Taxes [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return 0, 0, fmt.Errorf("maxium retries reached")
}

func CreateOrder(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, product ProductDetails, customer CustomerAddress,
	proxyUrl string, shippingCost, taxAmount, totalAmount int) (OrderDetails, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		spuId, err := strconv.ParseInt(product.SpuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		skuId, err := strconv.ParseInt(product.SkuId, 10, 64)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Convert String To Int, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		ordered := OrderedMap{
			{"userId", customer.UserId},
			{"addressId", customer.AddressId},
			{"totalAmount", product.Price * product.Quantity},
			{"orderTotalAmount", totalAmount + taxAmount + shippingCost},
			{"skuItem", []OrderedMap{
				{
					{"spuId", spuId},
					{"skuId", skuId},
					{"count", product.Quantity},
					{"skuCount", product.Quantity},
					{"price", product.Price},
					{"title", task.Size},
					{"spuTitle", product.ProductName},
					{"discountPrice", product.Price},
					{"currentSKUInCartNum", product.Quantity},
				},
			}},
			{"discountCode", nil},
			{"userCouponID", nil},
			{"mpUserCouponID", nil},
			{"activityId", nil},
			{"giftId", nil},
			{"express", OrderedMap{
				{"code", "STANDARD"},
				{"name", "Standard"},
				{"price", shippingCost},
			}},
			{"billAddressId", customer.AddressId},
			{"orderCreatePage", 1},
			{"snapshotID", ""},
			{"taxAmount", taxAmount},
			{"gwcClickID", ""},
			{"gwcProvider", ""},
			{"activities", []string{}},
			{"trafficSource", ""},
			{"trafficPlatform", ""},
			{"megaClotSpecialType", ""},
			{"currency", "USD"},
			{"isBox", false},
			{"captcha_data", nil},
		}

		data, err := ordered.MarshalJSON()
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/shop/v1/shop/placeOrderMix", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Verbose(fmt.Sprintf("Task %s: Creating Popmart Order", task.TaskId))
		jsonPayload, err := json.Marshal(CreatePayload{
			UserId:           customer.UserId,
			AddressId:        customer.AddressId,
			TotalAmount:      product.Price * product.Quantity,
			OrderTotalAmount: totalAmount + taxAmount + shippingCost,
			SkuItem: []CreateItem{
				{
					SpuId:         spuId,
					SkuId:         skuId,
					Count:         product.Quantity,
					SkuCount:      product.Quantity,
					Price:         product.Price,
					Title:         task.Size,
					SpuTitle:      product.ProductName,
					DiscountPrice: product.Price,
					CurrentCart:   product.Quantity,
				},
			},
			DiscountCode:   nil,
			UserCouponId:   nil,
			MpUserCouponId: nil,
			ActivityId:     nil,
			GiftId:         nil,
			Express: ExpressData{
				Code:  "STANDARD",
				Name:  "Standard",
				Price: shippingCost,
			},
			BillAddressId:       customer.AddressId,
			OrderCreatePage:     1,
			SnapshotId:          "",
			TaxAmount:           taxAmount,
			GwcClickID:          "",
			GwcProvider:         "",
			Activities:          []string{},
			TrafficSource:       "",
			TrafficPlatform:     "",
			MegaClotSpecialType: "",
			Currency:            "USD",
			IsBox:               false,
			Captcha:             nil,
			S:                   tdResp.S,
			T:                   int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/shop/v1/shop/placeOrderMix", strings.NewReader(string(jsonPayload)))
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
			"td-session-path":    {"/shop/v1/shop/placeOrderMix"},
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

		logger.Info(string(respBody))

		var createResp CreateResp
		if err := json.Unmarshal(respBody, &createResp); err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			switch createResp.Message {
			case "success":
				orderDetails := OrderDetails{
					ProductName:  product.ProductName,
					ProductImage: product.MainImage,
					SkuId:        product.SkuId,
					SpuId:        product.SpuId,
					ProductPrice: int64(product.Price),
					TotalAmount:  int64(createResp.Data.Amount.Value),
					OrderNumber:  createResp.Data.OrderNo,
				}
				return orderDetails, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Creating Popmart Order [%s], Retrying [%d]", task.TaskId, createResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Creating Popmart Order [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return OrderDetails{}, fmt.Errorf("maxium retries reached")
}

func ProcessPayment(task helpers.Task, logger *helpers.ColorizedLogger, client tls_client.HttpClient, userData helpers.UserData, order OrderDetails, accountEmail, proxyUrl, checkoutAttemptId string) (helpers.Webhook, error) {
	for retryCount := range make([]struct{}, helpers.MaxRetries) {
		ms := time.Now().UnixNano() / int64(time.Millisecond)
		payMark := strconv.FormatInt(ms, 10)

		adyenData, err := AdyenHelper(logger, task, order, checkoutAttemptId)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Encode Adyen Data, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		removeNonDigits := func(input string) string {
			re := regexp.MustCompile(`\D`)
			return re.ReplaceAllString(input, "")
		}
		cardNumber := removeNonDigits(task.Profile.CardNumber)

		ordered := OrderedMap{
			{"payType", "dropIn"},
			{"orderNo", order.OrderNumber},
			{"payMark", payMark},
			{"platform", "adyen"},
			{"cardInfo", OrderedMap{
				{"lastFour", cardNumber[len(cardNumber)-4:]},
				{"cardBin", cardNumber[:6]},
				{"holderName", task.Profile.Name},
			}},
			{"adyen", adyenData},
		}

		data, err := ordered.MarshalJSON()
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal API Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		tdResp, err := api.TD(logger, data, task.TaskId, "/shop/v1/shop/cash/desk/adyen/pay", proxyUrl, "post", helpers.UserAgent)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Generate TD Parameters, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		logger.Warn(fmt.Sprintf("Task %s: Processing Payment", task.TaskId))
		jsonPayload, err := json.Marshal(PaymentPayload{
			PayType:  "dropIn",
			OrderNo:  order.OrderNumber,
			PayMark:  payMark,
			Platform: "adyen",
			Info: CardInfo{
				LastFour:   cardNumber[len(cardNumber)-4:],
				CardBin:    cardNumber[:6],
				HolderName: task.Profile.Name,
			},
			Adyen: adyenData,
			S:     tdResp.S,
			T:     int64(tdResp.T),
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Marshal Request Payload, Retrying [%d]", task.TaskId, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}

		req, err := http.NewRequest("POST", "https://prod-na-api.popmart.com/shop/v1/shop/cash/desk/adyen/pay", strings.NewReader(string(jsonPayload)))
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
			"td-session-path":    {"/shop/v1/shop/cash/desk/adyen/pay"},
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

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode <= 299:
			var processResp ProcessResp
			if err := json.Unmarshal(respBody, &processResp); err != nil {
				logger.Error(fmt.Sprintf("Task %s: Failed To Unmarshal JSON Response Body, Retrying [%d]", task.TaskId, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}

			switch processResp.Message {
			case "success":
				webhook := helpers.Webhook{
					Type:        "Success",
					Account:     accountEmail,
					Site:        task.Site,
					Mode:        task.Mode,
					Product:     order.ProductName,
					Size:        task.Size,
					OrderNumber: order.OrderNumber,
					Profile:     task.Profile.ProfileName,
					ProxyGroup:  task.ProxyGroup,
					Image:       order.ProductImage,
				}

				helpers.IncrementCheckedOut()
				logger.Silly(fmt.Sprintf("Task %s: Successful Checkout 🌙", task.TaskId))
				return webhook, nil
			case "This transaction is risky. Please use another card or payment method.":
				webhook := helpers.Webhook{
					Type:        "Failure",
					Account:     accountEmail,
					Site:        task.Site,
					Mode:        task.Mode,
					Product:     order.ProductName,
					Size:        task.Size,
					OrderNumber: order.OrderNumber,
					Profile:     task.Profile.ProfileName,
					ProxyGroup:  task.ProxyGroup,
					Image:       order.ProductImage,
				}
				return webhook, nil
			case "The transaction was declined or flagged as risky. Please check with your bank.":
				webhook := helpers.Webhook{
					Type:        "Failure",
					Account:     accountEmail,
					Site:        task.Site,
					Mode:        task.Mode,
					Product:     order.ProductName,
					Size:        task.Size,
					OrderNumber: order.OrderNumber,
					Profile:     task.Profile.ProfileName,
					ProxyGroup:  task.ProxyGroup,
					Image:       order.ProductImage,
				}
				return webhook, nil
			default:
				logger.Error(fmt.Sprintf("Task %s: Error Processing Payment [%s], Retrying [%d]", task.TaskId, processResp.Message, retryCount+1))
				helpers.Delay(task.Delay)
				retryCount++
				continue
			}
		default:
			logger.Error(fmt.Sprintf("Task %s: Error Processing Payment [%d], Retrying [%d]", task.TaskId, resp.StatusCode, retryCount+1))
			helpers.Delay(task.Delay)
			retryCount++
			continue
		}
	}

	logger.Error(fmt.Sprintf("Task %s: Max Retries Has Been Reached", task.TaskId))
	return helpers.Webhook{}, fmt.Errorf("maxium retries reached")
}
