## Todos

allow init bash scripts in version zips

### Low Prio

make the acceptance tests only use high level functions like: goToChangePasswordPage, enterOldAndNewPassword etc.

user "sample" should always have the same cookie, so that I dont need to login every time when working with it locally
-> add a test the backend in native mode has fixed sample key, while in prod mode, the key changes when doing a second login


when updating, I am requesting all version of an app, which might be overkill; if necessary, optimize performance -> see GetVersionsHandler

for later: I do not like that .env approach, that backend fails to start if the .env is not present -> rather do the welcome page approach?
