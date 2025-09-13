TODO

* test case: two users register with same email address, first user validated is accepted, second one is denied (email already exists)
* when creating account, mention that other users using the store cant see the email address
  * or rather get rid of email?
  * when an admin uploads a new configuration of email settings, an automatic connection should be made to check whether it works; if not, email config is denied
* maybe make an implementation for user registration/validation like: if test mode -> return static sample registration code, otherwise create a random one
* introduce deepstack error wrapping in repos
* remove "log/respond" duplication in handlers
* replace repo tests by component tests to not leave out important use cases in test suite
* get rid of "data" folders (using same logger approach as in cloud)
* two deployments:
  * local, app store + database
  * prod, same + watchtotwer + traefik
* deployment: build image locally and test it; if passing, push image to dockerhub and download to hetzner server for deployment
* get rid of all panics and exit(1)
* replace "http.Error(w," with new logging+response system
* ged rid of native mode, only run in docker containers, have a test and prod profile config
* introduce unit tests, mocks, wire etc; shift business logic to units
* application should not fail on first boot when .env file is not there; in such case, simply generate a template -> make a warning log, that this still needs configuration
* I also want here that unit tests detect compile errors in modules with build tags and tests like in "cloud"
* user should be able to see the sizes of his version when listed
* make an API call like updateApps which delivers a list of apps, and is responded with just the new or updated apps (one total request is more efficient than one request per app)
* add anubis proxy to protect the server against crawlers etc? or maybe crowdsec?
 i dont like .env file approach -> idea: create a single admin account, with static name "admin"; for the "apps updater" CLI tool to here and also add administrative API to it. for example: set email settings, disable/ban user, reset user password etc.
  * replace GUI with that CLI tools
  * add acceptance tests by having tests using the CLI tools and asserting the output
* there is some security logic in the handlers for now, which should be in a service instead
* vai unit tests, check that a user cant upload a version > 1 MB, or apps with total size of > 10 MB
* add option to change email address; maybe make a field like "was email verified"?

### frontend 

* replace by CLI tool entirely; introduce cobra, so one cobra command runs the app store backend, another interacts with it, which makes sharing the library very easy
* cloud should import app store in order to get access to app store client, so move that from "shared" to this repo