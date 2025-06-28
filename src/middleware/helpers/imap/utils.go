package imap

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	helpers "popmart/src/middleware/helpers"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func ExtractPopmartCode(htmlContent string) (string, error) {
	pattern := `(?i)<div[^>]*>\s*(\d{6})\s*</div>`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(htmlContent)
	if len(match) > 1 {
		return match[1], nil
	}

	return "", fmt.Errorf("no verification code found in email body")
}

func ImapConnection(logger *helpers.ColorizedLogger, taskId, email, imapEmail, imapPassword string) (string, error) {
	var host string
	switch {
	case strings.Contains(imapEmail, "gmail.com"):
		host = "imap.gmail.com"
	case strings.Contains(imapEmail, "outlook.com"):
		host = "outlook.office365.com"
	case strings.Contains(imapEmail, "icloud.com"):
		if !strings.Contains(imapEmail, "@") {
			return "", fmt.Errorf("invalid iCloud email format: must include '@'")
		}
		host = "imap.mail.me.com"
	case strings.Contains(imapEmail, "thexyzstore.com"):
		host = "imap.thexyzstore.com"
	default:
		return "", fmt.Errorf("no supported email domains")
	}

	c, err := client.DialTLS(fmt.Sprintf("%s:993", host), nil)
	if err != nil {
		return "", fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer func() {
		if cerr := c.Logout(); cerr != nil {
			logger.Error(fmt.Sprintf("Task %s: Error Logging Out Of IMAP Server: %v", taskId, cerr))
		}
	}()

	if err := c.Login(imapEmail, imapPassword); err != nil {
		return "", fmt.Errorf("failed to login: %v", err)
	}

	timeout := time.After(45 * time.Second)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			logger.Error(fmt.Sprintf("Task %s: Timeout reached, Logging Out After 45 Seconds", taskId))
			return "", nil
		case <-ticker.C:
			code, err := CheckForEmails(c, logger, taskId, email)
			if err != nil {
				logger.Error(fmt.Sprintf("Task %s: Error Checking Emails: %v", taskId, err))
				continue
			}
			if code != "" {
				return code, nil
			}
		}
	}
}

func CheckForEmails(c *client.Client, logger *helpers.ColorizedLogger, taskId, email string) (string, error) {
	folders := []string{"INBOX", "Junk", "Spam"}
	icloudRegex := regexp.MustCompile(`^na\.support_at_popmart_com_[a-zA-Z0-9_]+@icloud\.com$`)

	for _, folder := range folders {
		_, err := c.Select(folder, false)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Select Folder %s: %v", taskId, folder, err))
			continue
		}

		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{imap.SeenFlag}
		criteria.Header.Add("To", email)

		if !strings.Contains(email, "icloud.com") {
			criteria.Header.Add("From", "na.support@popmart.com")
		}

		ids, err := c.Search(criteria)
		if err != nil {
			logger.Error(fmt.Sprintf("Task %s: Failed To Select Folder %s: %v", taskId, folder, err))
			continue
		}
		if len(ids) == 0 {
			continue
		}

		seqSet := new(imap.SeqSet)
		seqSet.AddNum(ids...)

		section := &imap.BodySectionName{}
		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)

		go func() {
			done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages)
		}()

		for msg := range messages {
			if msg == nil || msg.Envelope == nil {
				logger.Warn(fmt.Sprintf("Task %s: Skipping Message With Nil Envelope", taskId))
				continue
			}

			body := msg.GetBody(section)
			if body == nil {
				logger.Warn(fmt.Sprintf("Task %s: Message Body Is Empty, Skipping", taskId))
				continue
			}

			mr, err := mail.CreateReader(body)
			if err != nil {
				logger.Error(fmt.Sprintf("Task %s: Failed To Create Email Reader: %v", taskId, err))
				continue
			}

			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					logger.Error(fmt.Sprintf("Task %s: Error Reading Email Part: %v", taskId, err))
					continue
				}

				switch part.Header.(type) {
				case *mail.InlineHeader:
					content, _ := io.ReadAll(part.Body)
					if strings.Contains(email, "icloud.com") {
						fromAddress := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
						if !icloudRegex.MatchString(fromAddress) {
							logger.Warn(fmt.Sprintf("Task %s: Sender %s Does Not Match ICloud Regex, Skipping", taskId, fromAddress))
							continue
						}
					}

					code, err := ExtractPopmartCode(string(content))
					if err == nil {
						item := imap.FormatFlagsOp(imap.AddFlags, true)
						flags := []any{imap.SeenFlag}
						err = c.Store(seqSet, item, flags, nil)
						if err != nil {
							logger.Error(fmt.Sprintf("Task %s: Failed To Mark Email As Seen: %v", taskId, err))
						}
						return code, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no new Nike emails found")
}
