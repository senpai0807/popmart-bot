package desktop

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	helpers "popmart/src/middleware/helpers"
	discord "popmart/src/middleware/helpers/discord"
)

func PopmartDesktop(task helpers.Task, logger *helpers.ColorizedLogger) {
	logger.Info(fmt.Sprintf("Task %s: Starting Popmart %s Task", task.TaskId, task.Mode))
	Desktop(task, logger)
}

func Desktop(task helpers.Task, logger *helpers.ColorizedLogger) {
	if len(task.Proxies) == 0 {
		logger.Error(fmt.Sprintf("Task %s: No Proxies Are Available", task.TaskId))
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rawProxy := task.Proxies[rng.Intn(len(task.Proxies))]

	parts := strings.Split(rawProxy, ":")
	if len(parts) != 4 {
		logger.Error(fmt.Sprintf("Task %s: Invalid Proxy Format: %s", task.TaskId, rawProxy))
		return
	}

	ip, port, user, pass := parts[0], parts[1], parts[2], parts[3]
	proxyURL := fmt.Sprintf("http://%s:%s@%s:%s", user, pass, ip, port)

	logger.Verbose(fmt.Sprintf("Task %s: Using Proxy - %s", task.TaskId, proxyURL))
	logger.Verbose(fmt.Sprintf("Task %s: Creating Request Client", task.TaskId))
	client, err := helpers.CreateTLSClient(proxyURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Create Request Client: %v", task.TaskId, err))
		return
	}

	accountParts := strings.SplitN(task.Account, ":", 2)
	if len(accountParts) != 2 {
		logger.Error(fmt.Sprintf("Task %s: Invalid Account Format: %s", task.TaskId, task.Account))
		return
	}

	accountEmail := accountParts[0]
	accountPassword := accountParts[1]
	logger.Verbose(fmt.Sprintf("Task %s: Using Account - %s", task.TaskId, accountEmail))

	var userData helpers.UserData
	logger.Verbose(fmt.Sprintf("Task %s: Checking For Existing Session For %s", task.TaskId, accountEmail))

	userData, err = helpers.FetchSession(logger, task.TaskId, accountEmail)
	if err != nil {
		logger.Warn(fmt.Sprintf("Task %s: No Session Was Found For %s, Logging In", task.TaskId, accountEmail))
		checkErr := CheckExists(task, logger, client, accountEmail, proxyURL)
		if checkErr != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Check Account Existence: %s", task.TaskId, task.Account))
			return
		}

		helpers.Delay(task.Delay)
		userData, err = Login(task, logger, client, accountEmail, accountPassword, proxyURL)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Log Into Popmart Account", task.TaskId))
			return
		}
	}

	helpers.Delay(task.Delay)
	productDetails, err := FetchProduct(task, logger, client, proxyURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Fetch Product Details", task.TaskId))
		return
	}

	helpers.Delay(task.Delay)
	var customerAddress CustomerAddress
	customerAddress, err = FetchAddress(task, logger, client, userData, proxyURL)
	if err != nil {
		customerAddress, err = AddAddress(task, logger, client, userData, proxyURL)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Submit Shipping Information", task.TaskId))
			return
		}
	}

	helpers.Delay(task.Delay)
	atcErr := AddToCart(task, logger, client, userData, productDetails, proxyURL)
	if atcErr != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Add Product To Cart", task.TaskId))
		return
	}

	helpers.Delay(task.Delay)
	shippingCost, err := FetchRates(task, logger, client, userData, productDetails, proxyURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Fetch Shipping Rates", task.TaskId))
		return
	}

	helpers.Delay(task.Delay)
	taxAmount, totalAmount, err := CalculateTaxes(task, logger, client, userData, productDetails, customerAddress, proxyURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Calculate Taxes", task.TaskId))
		return
	}

	helpers.Delay(task.Delay)
	orderDetails, err := CreateOrder(task, logger, client, userData, productDetails, customerAddress, proxyURL, shippingCost, taxAmount, totalAmount)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Create Popmart Order", task.TaskId))
		return
	}

	var checkoutErr error

	switch task.Payment {
	case "Card":
		helpers.Delay(task.Delay)
		checkoutAttemptId, err := FetchCheckoutId(task, logger, client, proxyURL)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Fetch Checkout Attempt ID", task.TaskId))
			return
		}

		helpers.Delay(task.Delay)
		webhookData, err := ProcessPayment(task, logger, client, userData, orderDetails, accountEmail, proxyURL, checkoutAttemptId)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Process Payment", task.TaskId))
			return
		}

		checkoutErr = discord.SendWebhook(logger, webhookData, task.TaskId)
	case "Paypal":
		helpers.Delay(task.Delay)
		webhookData, err := Paypal(task, logger, client, userData, orderDetails, accountEmail, proxyURL)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Create Paypal Checkout Link", task.TaskId))
			return
		}

		checkoutErr = discord.SendPaypal(logger, webhookData, task.TaskId)
	default:
		logger.Error(fmt.Sprintf("Task %s: Unsupported Payment Type Selected", task.TaskId))
		return
	}

	if checkoutErr != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Process Popmart Order", task.TaskId))
		return
	}
}
