package tools

import (
	"sync"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

// TODO !! make this persisting to database; i can store hashed registration codes and passwords salted-hash in database; so that this survives reboots
type EmailVerifierImpl struct {
	WaitingList sync.Map
}

func (e *EmailVerifierImpl) Store(code string, form *store.RegistrationForm) {
	e.WaitingList.Store(code, form)
}

func (e *EmailVerifierImpl) Load(code string) (*store.RegistrationForm, error) {
	value, ok := e.WaitingList.Load(code)
	if !ok {
		return nil, u.Logger.NewError("code not found")
	}

	form, ok := value.(*store.RegistrationForm)
	if !ok {
		return nil, u.Logger.NewError("invalid type for registration form")
	}
	return form, nil
}

func (e *EmailVerifierImpl) Delete(code string) {
	e.WaitingList.Delete(code)
}

func (e *EmailVerifierImpl) Clear() {
	e.WaitingList = sync.Map{}
}
