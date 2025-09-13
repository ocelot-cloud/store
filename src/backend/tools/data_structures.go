package tools

// TODO !! centrally collect data structures here

var (
	SampleUser     = "samplemaintainer"
	SampleApp      = "sampleapp"
	SampleVersion  = "0.0.1"
	SampleEmail    = "sample@sample.com"
	SamplePassword = "samplepassword"
)

type User struct {
	Id                int
	Name              string
	Email             string
	HashedPassword    string
	HashedCookieValue *string
	ExpirationDate    *string
	UsedSpaceInBytes  int
}

type App struct {
	Id      int
	OwnerId int
	Name    string
}
