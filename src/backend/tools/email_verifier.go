package tools

import (
	"fmt"
	"sync"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type EmailVerifierImpl struct {
	WaitingList sync.Map
}

func (e *EmailVerifierImpl) Store(code string, form *store.RegistrationForm) {
	e.WaitingList.Store(code, form)
}

func (e *EmailVerifierImpl) Load(code string) (*store.RegistrationForm, error) {
	value, ok := e.WaitingList.Load(code)
	if !ok {
		return nil, fmt.Errorf("code not found")
	}

	form, ok := value.(*store.RegistrationForm)
	if !ok {
		u.Logger.Error("Invalid type for registration form")
		return nil, fmt.Errorf("invalid type for registration form")
	}
	return form, nil
}

func (e *EmailVerifierImpl) Delete(code string) {
	e.WaitingList.Delete(code)
}

func (e *EmailVerifierImpl) Clear() {
	e.WaitingList = sync.Map{}
}
