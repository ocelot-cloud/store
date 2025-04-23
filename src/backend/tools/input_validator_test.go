package tools

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestValidateName(t *testing.T) {
	assert.Nil(t, Validate("validusername", User))
	assert.Nil(t, Validate("user123", User))
	assert.NotNil(t, Validate("InvalidUsername", User))          // Contains uppercase
	assert.NotNil(t, Validate("user!@#", User))                  // Contains special characters
	assert.NotNil(t, Validate("us", User))                       // Too short
	assert.NotNil(t, Validate("thisusernameiswaytoolong", User)) // Too long
}

func TestValidateVersion(t *testing.T) {
	assert.Nil(t, Validate("valid.versionname", VersionType))
	assert.Nil(t, Validate("version123", VersionType))
	assert.Nil(t, Validate("version.name123", VersionType))
	assert.NotNil(t, Validate("invalid.versionname!", VersionType))             // Contains special characters other than dot
	assert.NotNil(t, Validate("ta", VersionType))                               // Too short
	assert.NotNil(t, Validate("this.versionname.is.way.too.long", VersionType)) // Too long
}

func TestValidatePassword(t *testing.T) {
	assert.Nil(t, Validate("validpassword!", Password))
	assert.Nil(t, Validate("valid_pass123", Password))
	assert.Nil(t, Validate("InvalidPassword", Password)) // Contains uppercase
	assert.Nil(t, Validate("valid!@#", Password))        // Contains special characters
	assert.NotNil(t, Validate("1234567", Password))      // Too short
	assert.Nil(t, Validate("12345678", Password))
	assert.NotNil(t, Validate("thispasswordiswaytoolong_xxxxx!", Password)) // Too long
}

func TestValidateCookie(t *testing.T) {
	sixtyOneHexDecimalLetters := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"

	assert.NotNil(t, Validate(sixtyOneHexDecimalLetters, Cookie))
	assert.Nil(t, Validate(sixtyOneHexDecimalLetters+"f", Cookie))
	assert.NotNil(t, Validate(sixtyOneHexDecimalLetters+"ff", Cookie))
	assert.NotNil(t, Validate(sixtyOneHexDecimalLetters+"g", Cookie))
	assert.NotNil(t, Validate("", Cookie))
}

func TestValidateEmail(t *testing.T) {
	assert.Nil(t, Validate("admin@admin.com", Email))
	assert.NotNil(t, Validate("@admin.com", Email))
	assert.NotNil(t, Validate("admin@.com", Email))
	assert.NotNil(t, Validate("admin@admin.", Email))
	assert.NotNil(t, Validate("adminadmin.com", Email))
	assert.NotNil(t, Validate("admin@admincom", Email))

	thirtyCharacters := "abcdefghijklmnopqrstuvwxyz1234"
	validEmail := fmt.Sprintf("%s@%s.de", thirtyCharacters, thirtyCharacters)
	assert.Nil(t, Validate(validEmail, Email))
	tooLongEmail := fmt.Sprintf("%s@%s.com", thirtyCharacters, thirtyCharacters)
	assert.NotNil(t, Validate(tooLongEmail, Email))
}

func TestValidateNumber(t *testing.T) {
	assert.Nil(t, Validate("0", Number))
	assert.Nil(t, Validate("1", Number))
	assert.NotNil(t, Validate("-1", Number))
	assert.NotNil(t, Validate("a", Number))
	assert.NotNil(t, Validate("A", Number))
	assert.NotNil(t, Validate("z", Number))
	assert.NotNil(t, Validate("Z", Number))
	assert.NotNil(t, Validate("-", Number))
	assert.NotNil(t, Validate("_", Number))
	assert.NotNil(t, Validate(".", Number))
	assert.NotNil(t, Validate(",", Number))

	twentyDigitNumber := "01234567890123456789"
	assert.Nil(t, Validate(twentyDigitNumber, Number))
	assert.NotNil(t, Validate(twentyDigitNumber+"0", Number))
}

func TestSearchTerm(t *testing.T) {
	assert.Nil(t, Validate("", SearchTerm))
	assert.Nil(t, Validate("a", SearchTerm))
	assert.Nil(t, Validate("1", SearchTerm))
	assert.Nil(t, Validate("0123456789abcdefghij", SearchTerm))
	assert.NotNil(t, Validate("asdf!", SearchTerm))
}
