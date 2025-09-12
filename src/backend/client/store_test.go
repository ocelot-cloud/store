package store

import (
	"testing"
)

func TestAppStoreClientInterfaceImplementation(t *testing.T) {
	var _ AppStoreClient = &AppStoreClientImpl{}
}
