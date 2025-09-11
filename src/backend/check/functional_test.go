//go:build component

package check

import (
	"testing"

	"github.com/ocelot-cloud/deepstack"
)

// TODO !! duplication with cloud -> move to "shared"
func AssertDeepStackErrorWithCode(t *testing.T, err error, expectedResponseBodyErrorMessage string, expectedStatusCode int) {
	deepstack.AssertDeepStackError(t, err, "request failed", "response_body", expectedResponseBodyErrorMessage, "status_code", expectedStatusCode)
}
