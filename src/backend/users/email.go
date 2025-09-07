package users

import (
	"bufio"
	"fmt"
	"ocelot/store/tools"
	"os"
	"strconv"
	"strings"

	"github.com/ocelot-cloud/deepstack"
	u "github.com/ocelot-cloud/shared/utils"
	"gopkg.in/gomail.v2"
)

// TODO !! get rid of .env file
const envFilePath = "data/.env"

// TODO !! global var
var (
	SMTP_PORT                                          int
	HOST, SMTP_HOST, EMAIL, EMAIL_USER, EMAIL_PASSWORD string
)

func InitializeEnvs() error {
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		defaultEnv := []byte("HOST=http://localhost:8082\nSMTP_HOST=smtps.sample.com\nSMTP_PORT=465\nEMAIL=sample@sample.com\nEMAIL_USER=sample\nEMAIL_PASSWORD=password\n")
		err = os.WriteFile(envFilePath, defaultEnv, 0600)
		if err != nil {
			u.Logger.Error("Failed to create .env file", deepstack.ErrorField, err)
			return fmt.Errorf("failed to create .env file")
		}
		u.Logger.Info(".env file created with default values")
		return nil
	} else {
		var file *os.File
		file, err = os.Open(envFilePath)
		if err != nil {
			u.Logger.Error("Failed to open .env file", deepstack.ErrorField, err)
			os.Exit(1)
		}
		defer u.Close(file)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				err = os.Setenv(parts[0], parts[1])
				if err != nil {
					u.Logger.Error("failed to set environment variable", tools.EnvVarField, parts[0], deepstack.ErrorField, err)
					os.Exit(1)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			u.Logger.Error("Error reading .env file", deepstack.ErrorField, err)
			os.Exit(1)
		}

		HOST = GetEnv("HOST")
		SMTP_HOST = GetEnv("SMTP_HOST")
		smtpPort := GetEnv("SMTP_PORT")
		SMTP_PORT, err = strconv.Atoi(smtpPort)
		if err != nil {
			u.Logger.Error("Failed to parse SMTP_PORT env", tools.SmtpPortEnvField, smtpPort, deepstack.ErrorField, err)
			os.Exit(1)
		}
		EMAIL = GetEnv("EMAIL")
		EMAIL_USER = GetEnv("EMAIL_USER")
		EMAIL_PASSWORD = GetEnv("EMAIL_PASSWORD")

		u.Logger.Info(".env file loaded successfully")
		return err
	}
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		// TODO this was a panic error before, but not sure whether it should be panic
		u.Logger.Error("environment variable not set")
		return ""
	} else {
		u.Logger.Debug("Loaded env", tools.EnvVarField, key)
		return value
	}
}

func sendVerificationEmail(to, code string) error {
	if tools.UseMailMockClient {
		u.Logger.Debug("Mock email client used, not sending email")
		return nil
	} else {
		verificationLink := HOST + "/validate?code=" + code
		m := gomail.NewMessage()
		m.SetHeader("From", EMAIL)
		m.SetHeader("To", to)
		m.SetHeader("Subject", "Verify Your Email Address")
		m.SetBody("text/html", fmt.Sprintf("<p>Please verify your email address by clicking the following link to complete your registration for the Ocelot App Store:</p><p><a href='%s'>Verify Email</a></p>", verificationLink))
		d := gomail.NewDialer(SMTP_HOST, SMTP_PORT, EMAIL_USER, EMAIL_PASSWORD)
		u.Logger.Debug("Sending validation email to", tools.TargetEmailField, to)
		return d.DialAndSend(m)
	}
}
