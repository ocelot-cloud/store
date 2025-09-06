TODO

* sometimes the application does not start due to database interference (should be fixed when there is only one database file left)

```
  tux@tux:~/Dokumente/workspace/store/src/ci-runner$ go build && ./ci-runner test all
  in directory './backend', executing 'rm -rf data'
  => Command successful. Time taken: 0.002 seconds.

==== Testing units ====

in directory './backend/docker', executing 'docker compose -f docker-compose-dev.yml up -d'
Container ocelotcloud_store_postgres  Creating
Container ocelotcloud_store_postgres  Error response from daemon: Conflict. The container name "/ocelotcloud_store_postgres" is already in use by container "b70e97f3755bfd8c63d92fd6fd23a01ff734add5508250e97791b4cc6b0d46bc". You have to remove (or rename) that container to be able to reuse that name.
Error response from daemon: Conflict. The container name "/ocelotcloud_store_postgres" is already in use by container "b70e97f3755bfd8c63d92fd6fd23a01ff734add5508250e97791b4cc6b0d46bc". You have to remove (or rename) that container to be able to reuse that name.
=> Command failed. Time taken: 0.078 seconds.. Error: exit status 1

cleanup method called
calling custom cleanup function
```

* unit tests should fail if there is a compile error in the component tests
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