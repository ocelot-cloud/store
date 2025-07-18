package users

import (
	"bufio"
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"gopkg.in/gomail.v2"
	"ocelot/store/tools"
	"os"
	"strconv"
	"strings"
)

const envFilePath = "data/.env"

var (
	SMTP_PORT                                          int
	HOST, SMTP_HOST, EMAIL, EMAIL_USER, EMAIL_PASSWORD string
)

func InitializeEnvs() error {
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		defaultEnv := []byte("HOST=http://localhost:8082\nSMTP_HOST=smtps.sample.com\nSMTP_PORT=465\nEMAIL=sample@sample.com\nEMAIL_USER=sample\nEMAIL_PASSWORD=password\n")
		err := os.WriteFile(envFilePath, defaultEnv, 0600)
		if err != nil {
			tools.Logger.ErrorF("Failed to create .env file: %v", err)
			return fmt.Errorf("failed to create .env file")
		}
		tools.Logger.InfoF(".env file created with default values. Exiting.")
		return fmt.Errorf("default .env file created, please set values and start application again")
	} else {
		file, err := os.Open(envFilePath)
		if err != nil {
			tools.Logger.ErrorF("Failed to open .env file: %v", err)
			os.Exit(1)
		}
		defer utils.Close(file)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				err = os.Setenv(parts[0], parts[1])
				if err != nil {
					tools.Logger.ErrorF("failed to set environment variable %s: %v", parts[0], err)
					os.Exit(1)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			tools.Logger.ErrorF("Error reading .env file: %v", err)
			os.Exit(1)
		}

		HOST = GetEnv("HOST")
		SMTP_HOST = GetEnv("SMTP_HOST")
		smtpPort := GetEnv("SMTP_PORT")
		SMTP_PORT, err = strconv.Atoi(smtpPort)
		if err != nil {
			tools.Logger.ErrorF("Failed to parse SMTP_PORT env with value '%s': %v", smtpPort, err)
			os.Exit(1)
		}
		EMAIL = GetEnv("EMAIL")
		EMAIL_USER = GetEnv("EMAIL_USER")
		EMAIL_PASSWORD = GetEnv("EMAIL_PASSWORD")

		tools.Logger.InfoF(".env file loaded successfully")
		return err
	}
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		// TODO this was a panic error before, but not sure whether it should be panic
		tools.Logger.ErrorF("environment variable not set")
		return ""
	} else {
		tools.Logger.DebugF("Loaded env %s", key)
		return value
	}
}

func sendVerificationEmail(to, code string) error {
	if tools.UseMailMockClient {
		tools.Logger.DebugF("Mock email client used, not sending email")
		return nil
	} else {
		verificationLink := HOST + "/validate?code=" + code
		m := gomail.NewMessage()
		m.SetHeader("From", EMAIL)
		m.SetHeader("To", to)
		m.SetHeader("Subject", "Verify Your Email Address")
		m.SetBody("text/html", fmt.Sprintf("<p>Please verify your email address by clicking the following link to complete your registration for the Ocelot App Store:</p><p><a href='%s'>Verify Email</a></p>", verificationLink))
		d := gomail.NewDialer(SMTP_HOST, SMTP_PORT, EMAIL_USER, EMAIL_PASSWORD)
		tools.Logger.DebugF("Sending validation email to %s", to)
		return d.DialAndSend(m)
	}
}
