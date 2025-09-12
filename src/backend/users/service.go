package users

import (
	"crypto/rand"
	"encoding/hex"
	"ocelot/store/tools"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type UserServiceImpl struct {
	Repo          UserRepository
	Config        *tools.Config
	EmailVerifier *tools.EmailVerifierImpl
}

func (r *UserServiceImpl) CreateUserAndReturnRegistrationCode(form *store.RegistrationForm) (string, error) {
	var key string
	// TODO !! quite implicit logic. Maybe a better option to say in test mode, when we create a user, his account needs no code for validation
	if r.Config.UseSampleDataForTesting {
		// TODO static sample key for testing, use from shared module
		key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	} else {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			u.Logger.Error("Failed to generate cookie", deepstack.ErrorField, err)
			return "", err
		}
		key = hex.EncodeToString(randomBytes)
	}
	u.Logger.Info("adding user to validation list", tools.UserField, form.User)
	r.EmailVerifier.Store(key, form)
	return key, nil
}
