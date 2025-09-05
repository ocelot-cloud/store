TODO

* ged rid of native mode, only run in docker containers, have a test and prod profile config
* introduce unit tests, mocks, wire etc; shift business logic to units
* application should not fail on first boot when .env file is not there; in such case, simply generate a template -> make a warning log, that this still needs configuration
* I also want here that unit tests detect compile errors in modules with build tags and tests like in "cloud"
* make mock that ignores the .env file in non-prod profiles
* make an API call like updateApps which delivers a list of apps, and is responded with just the new or updated apps (one total request is more efficient than one request per app)
* add anubis proxy to protect the server against crawlers etc? or maybe crowdsec?
 i dont like .env file approach -> idea: create a single admin account, with static name "admin"; for the "apps updater" CLI tool to here and also add administrative API to it. for example: set email settings, disable/ban user, reset user password etc.
  * replace GUI with that CLI tools
  * add acceptance tests by having tests using the CLI tools and asserting the output

### frontend 

* replace by CLI tool entirely; introduce cobra, so one cobra command runs the app store backend, another interacts with it, which makes sharing the library very easy