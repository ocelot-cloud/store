package users

import (
	"fmt"
	"ocelot/store/tools"

	u "github.com/ocelot-cloud/shared/utils"
	"gopkg.in/gomail.v2"
)

type EmailClientImpl struct {
	Config           *tools.Config
	EmailConfigStore *EmailConfigStoreImpl
}

func (e *EmailClientImpl) SendVerificationEmail(to, code string) error {
	if e.Config.UseMailMockClient {
		u.Logger.Debug("Mock email client used, not sending email")
		return nil
	} else {
		emailConfig, err := e.EmailConfigStore.GetEmailConfig()
		if err != nil {
			return err
		}
		verificationLink := emailConfig.AppStoreHost + "/validate?code=" + code
		m := gomail.NewMessage()
		m.SetHeader("From", emailConfig.EmailAddress)
		m.SetHeader("To", to)
		m.SetHeader("Subject", "Verify Your Email Address")
		m.SetBody("text/html", fmt.Sprintf("<p>Please verify your email address by clicking the following link to complete your registration for the Ocelot App Store:</p><p><a href='%s'>Verify Email</a></p>", verificationLink))
		d := gomail.NewDialer(emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.EmailAccountUsername, emailConfig.EmailAccountPassword)
		u.Logger.Debug("Sending validation email to", tools.TargetEmailField, to)
		return d.DialAndSend(m)
	}
}
