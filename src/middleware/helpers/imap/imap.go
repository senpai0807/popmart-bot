package imap

import (
	"fmt"

	helpers "popmart/src/middleware/helpers"
)

func FetchCode(logger *helpers.ColorizedLogger, taskId, email, imapEmail, imapPassword string) (string, error) {
	code, err := ImapConnection(logger, taskId, email, imapEmail, imapPassword)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Fetch Code From Email", taskId))
		return "", err
	}

	if code == "" {
		logger.Error(fmt.Sprintf("Task %s: No Verification Code Found In Inbox", taskId))
		return "", err
	}
	return code, nil
}
