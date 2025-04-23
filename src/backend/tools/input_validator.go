package tools

import (
	"fmt"
	"regexp"
)

type ValidationType int

const (
	User ValidationType = iota
	AppType
	VersionType
	Password
	Email
	Cookie
	Number
	ValidationCode
	SearchTerm
)

var validationTypeStrings = []string{"user", "app", "version", "password", "email", "cookie", "number", "code-validation", "search-term"}

func getValidationTypeString(validationType ValidationType) string {
	return validationTypeStrings[validationType]
}

var (
	namePattern           = regexp.MustCompile(`^[a-z0-9]{3,20}$`)
	versionPattern        = regexp.MustCompile(`^[a-z0-9.]{3,20}$`)
	passwordPattern       = regexp.MustCompile(`^[a-zA-Z0-9!@#$%&_,.?]{8,30}$`)
	emailPattern          = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	cookiePattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	numberPattern         = regexp.MustCompile(`^[0-9]{1,20}$`)
	codeValidationPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
	searchTermPattern     = regexp.MustCompile(`^[a-z0-9]{0,20}$`)
)

func Validate(input string, validationType ValidationType) error {
	var re *regexp.Regexp

	switch validationType {
	case User:
		re = namePattern
	case AppType:
		re = namePattern
	case VersionType:
		re = versionPattern
	case Password:
		re = passwordPattern
	case Email:
		re = emailPattern
	case Cookie:
		re = cookiePattern
	case Number:
		re = numberPattern
	case ValidationCode:
		re = codeValidationPattern
	case SearchTerm:
		re = searchTermPattern
	default:
		return fmt.Errorf("invalid validation type with index: %d", validationType)
	}

	if validationType == Email && len(input) > 64 {
		return fmt.Errorf("maximum email length of 64 characters is exceeded")
	}

	result := re.MatchString(input)
	if result {
		return nil
	} else {
		if validationType == Password || validationType == Cookie {
			Logger.Info("input validation failed for validation type: %s", getValidationTypeString(validationType))
		} else {
			Logger.Info("input validation failed for validation type '%s' with input '%s'", getValidationTypeString(validationType), input)
		}
		return fmt.Errorf("invalid signs or length of field: %s", getValidationTypeString(validationType))
	}
}
