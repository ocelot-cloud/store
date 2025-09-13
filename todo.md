TODO

* test case: two users register with same email address, first user validated is accepted, second one is denied (email already exists)
* when creating account, mention that other users using the store cant see the email address
  * or rather get rid of email?
  * when an admin uploads a new configuration of email settings, an automatic connection should be made to check whether it works; if not, email config is denied
* replace repo tests by component tests to not leave out important use cases in test suite
* user should be able to see the sizes of his versions when listed

* make an API call like updateApps which delivers a list of apps, and is responded with just the new or updated apps (one total request is more efficient than one request per app) -> uploading multiple versions at once with a single request
* via unit tests, check that a user cant upload a version > 1 MB, or apps with total size of > 10 MB
* add option to change email address -> do this in memory, if code is provided, then change email address

### frontend 

* replace by CLI tool entirely; introduce cobra, so one cobra command runs the app store backend, another interacts with it, which makes sharing the library very easy
* cloud should import app store in order to get access to app store client, so move that from "shared" to this repo

* in the end, deploy to hetzner server
  * deployment: build image locally and test it; if passing, push image to dockerhub and download to hetzner server for deployment