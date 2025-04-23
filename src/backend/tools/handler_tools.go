package tools

import (
	"encoding/json"
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"io"
	"net/http"
)

type ContextKey string

const UserCtxKey ContextKey = "user"

type ValidationJob struct {
	Value   string
	ValType ValidationType
}

func ReadBody[T any](r *http.Request) (*T, error) {
	var result T

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body: %w", err)
	}
	defer utils.Close(r.Body)

	if err = json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	var jobs []ValidationJob
	switch v := any(result).(type) {
	case App:
		jobs = []ValidationJob{
			{v.Maintainer, User},
			{v.Name, AppType},
		}
	case RegistrationForm:
		jobs = []ValidationJob{
			{v.User, User},
			{v.Password, Password},
			{v.Email, Email},
		}
	case utils.ChangePasswordForm:
		jobs = []ValidationJob{
			{v.OldPassword, Password},
			{v.NewPassword, Password},
		}
	case LoginCredentials:
		jobs = []ValidationJob{
			{v.User, User},
			{v.Password, Password},
		}
	case AppSearchRequest:
		jobs = []ValidationJob{
			{v.SearchTerm, SearchTerm},
		}
	}
	if err = ValidateJobs(jobs); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func ValidateJobs(jobs []ValidationJob) error {
	for _, job := range jobs {
		if err := Validate(job.Value, job.ValType); err != nil {
			return err
		}
	}
	return nil
}

func ReadBodyAsSingleString(w http.ResponseWriter, r *http.Request, validationType ValidationType) (string, error) {
	singleString, err := ReadBody[utils.SingleString](r)
	if err != nil {
		HandleInvalidInput(w, err)
		return "", fmt.Errorf("")
	}
	result := singleString.Value

	if err = Validate(result, validationType); err != nil {
		HandleInvalidInput(w, err)
		return "", fmt.Errorf("")
	}

	return result, nil
}

func HandleInvalidInput(w http.ResponseWriter, err error) {
	Logger.Info("invalid input: %v", err)
	http.Error(w, "invalid input", http.StatusBadRequest)
}

// GetUserFromContext Since only authenticated users are added to the context, it only works in protected handlers.
func GetUserFromContext(r *http.Request) string {
	return r.Context().Value(UserCtxKey).(string)
}
