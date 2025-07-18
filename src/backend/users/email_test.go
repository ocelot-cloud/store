package users

import (
	"github.com/ocelot-cloud/shared/assert"
	"os"
	"testing"
)

func DISABLED_TestSendMail(t *testing.T) {
	to := "sample@sample.com"
	assert.Nil(t, sendVerificationEmail(to, "1234"))
}

func TestInitializeEnv(t *testing.T) {
	if _, err := os.Stat(envFilePath); err == nil {
		err := os.Remove(envFilePath)
		assert.Nil(t, err)
	}

	err := InitializeEnvs()
	assert.Nil(t, err)

	_, err = os.Stat(envFilePath)
	assert.False(t, os.IsNotExist(err))

	err = InitializeEnvs()
	assert.Nil(t, err)

	assert.Equal(t, "http://localhost:8082", HOST)
	assert.Equal(t, "smtps.sample.com", SMTP_HOST)
	assert.Equal(t, 465, SMTP_PORT)
	assert.Equal(t, "sample@sample.com", EMAIL)
	assert.Equal(t, "sample", EMAIL_USER)
	assert.Equal(t, "password", EMAIL_PASSWORD)
}
