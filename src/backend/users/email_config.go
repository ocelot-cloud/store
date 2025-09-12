package users

import (
	"ocelot/store/tools"
	"strconv"

	u "github.com/ocelot-cloud/shared/utils"
)

// TODO !! to be tested
// TODO !! add handler for getting/setting email config + component tests -> use hardcoded password from env for auth?; Application should crash start when this variable is empty/not set

type EmailConfig struct {
	AppStoreHost         string
	SMTPHost             string
	SMTPPort             int
	EmailAddress         string
	EmailAccountUsername string
	EmailAccountPassword string
}

type EmailConfigStore interface {
	GetEmailConfig() (EmailConfig, error)
	SetEmailConfig(cfg EmailConfig) error
}

const (
	appStoreHost     = "APP_STORE_HOST"
	smtpHostKey      = "EMAIL_SMTP_HOST"
	smtpPortKey      = "EMAIL_SMTP_PORT"
	emailKey         = "EMAIL_ADDRESS"
	emailUserKey     = "EMAIL_ACCOUNT_USERNAME"
	emailPasswordKey = "EMAIL_ACCOUNT_PASSWORD"

	DefaultAppStoreHost  = "http://localhost"
	defaultSMTPHost      = "smtps.sample.com"
	defaultSMTPPort      = 465
	defaultEmail         = "sample@sample.com"
	defaultEmailUser     = "sample"
	defaultEmailPassword = "password"
)

type EmailConfigStoreImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
}

func (s *EmailConfigStoreImpl) GetEmailConfig() (*EmailConfig, error) {
	cfg := &EmailConfig{
		AppStoreHost:         DefaultAppStoreHost,
		SMTPHost:             defaultSMTPHost,
		SMTPPort:             defaultSMTPPort,
		EmailAddress:         defaultEmail,
		EmailAccountUsername: defaultEmailUser,
		EmailAccountPassword: defaultEmailPassword,
	}

	row, err := s.DatabaseProvider.GetDb().Query(`
		SELECT key, value
		FROM configs
		WHERE key IN ($1,$2,$3,$4,$5,$6)
	`, appStoreHost, smtpHostKey, smtpPortKey, emailKey, emailUserKey, emailPasswordKey)
	if err != nil {
		return nil, err
	}
	defer u.Close(row)

	m := map[string]string{}
	for row.Next() {
		var k, v string
		if err := row.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	if row.Err() != nil {
		return nil, row.Err()
	}

	if v, ok := m[appStoreHost]; ok {
		cfg.AppStoreHost = v
	}
	if v, ok := m[smtpHostKey]; ok {
		cfg.SMTPHost = v
	}
	if v, ok := m[smtpPortKey]; ok {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.SMTPPort = p
		}
	}
	if v, ok := m[emailKey]; ok {
		cfg.EmailAddress = v
	}
	if v, ok := m[emailUserKey]; ok {
		cfg.EmailAccountUsername = v
	}
	if v, ok := m[emailPasswordKey]; ok {
		cfg.EmailAccountPassword = v
	}

	return cfg, nil
}

func (s *EmailConfigStoreImpl) SetEmailConfig(cfg EmailConfig) error {
	if _, err := s.DatabaseProvider.GetDb().Exec(`
		INSERT INTO configs(key, value) VALUES
			($1,$2),($3,$4),($5,$6),($7,$8),($9,$10),($11,$12)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`,
		appStoreHost, cfg.AppStoreHost,
		smtpHostKey, cfg.SMTPHost,
		smtpPortKey, strconv.Itoa(cfg.SMTPPort),
		emailKey, cfg.EmailAddress,
		emailUserKey, cfg.EmailAccountUsername,
		emailPasswordKey, cfg.EmailAccountPassword,
	); err != nil {
		return err
	}

	return nil
}
