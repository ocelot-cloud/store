package users

import (
	"crypto/rand"
	"encoding/hex"
	"ocelot/store/tools"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type RegistrationCodeProvider struct {
	Config *tools.Config
}

func (p RegistrationCodeProvider) GenerateCode() (string, error) {
	if p.Config.UseSampleDataForTesting {
		return store.DefaultValidationCode, nil
	} else {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			return "", u.Logger.NewError(err.Error())
		}
		return hex.EncodeToString(randomBytes), nil
	}
}
