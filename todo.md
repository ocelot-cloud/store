TODO

* introduce unit tests, mocks, wire etc; shift business logic to units
* make mock that ignores the .env file in non-prod profiles
* make an API call like updateApps which delivers a list of apps, and is responded with just the new or updated apps (one total request is more efficient than one request per app)
* add anubis proxy to protect the server against crawlers etc? or maybe crowdsec?

### frontend 

* may be replaced by a CLI tool entirely
* merge cypress and frontend folder; use yarn as the package manager