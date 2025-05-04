let authCookie = ""

function getBaseUrl() {
    if (Cypress.env('CYPRESS_PROFILE') == "TEST") {
        return 'http://localhost:8081'
    } else {
        return 'http://localhost:8082'
    }
}

const baseUrl = getBaseUrl()
const hubWipePath = `http://localhost:8082/api/wipe-data`
const loginPath = `${baseUrl}/login`
const validationPath = `${baseUrl}/validate`
const registrationPath = `${baseUrl}/registration`
const changePasswordPath = `${baseUrl}/change-password`
const versionManagementPath = `${baseUrl}/versions`
const appsPath = `${baseUrl}/`
const maintainerName = "samplemaintainer"

const appName = "gitea"

function createApp() {
    cy.get('#input-app').type(appName);
    cy.get('#button-create-app').click();
    shallInvalidInputMessageBeShown(false, "app")
}

function shallInvalidInputMessageBeShown(shallBeShown: boolean, fieldName: string) {
    if (shallBeShown) {
        cy.get('body').should('contain.text', 'Invalid '+ fieldName);
    } else {
        cy.get('body').should('not.contain', 'Invalid ' + fieldName);
    }
}

function cancelAccountDeletion() {
    cy.get('#delete-account').click();
    cy.get('#button-close-popup').click();
    cy.get('#button-close-popup').should('not.exist');
    cy.get('#button-delete-confirmation').should('not.exist');
    cy.url().should('eq', appsPath);
}

function executeAccountDeletion() {
    cy.get('#delete-account').click();
    cy.get('#button-delete-confirmation').click();
    cy.get('#button-close-popup').should('not.exist');
    cy.get('#button-delete-confirmation').should('not.exist');
    cy.url().should('eq', loginPath);
}

function assertEmptyAppList() {
    cy.get('#app-list').find('.app-item').should('have.length', 0);
    cy.get('#button-delete-app').should('not.exist')
    cy.get('#button-edit-versions').should('not.exist')
}

function assertIsAppSelected(isSelected: boolean) {
    let prefix = ""
    if (!isSelected) {
        prefix = "not."
    }
    cy.get('#button-delete-app').should(prefix + 'exist')
    cy.get('#button-edit-versions').should(prefix + 'exist')
    cy.get('#app-list').find('.app-item').should('have.length', 1)
    cy.get('#selection-icon').should(prefix + 'exist')
}

function assertIsVersionSelected(isSelected: boolean) {
    let prefix = ""
    if (!isSelected) {
        prefix = "not."
    }

    cy.get('#button-download-version').should(prefix + 'exist')
    cy.get('#button-delete-version').should(prefix + 'exist')
    cy.get('#version-list').find('.version-item').should('have.length', 1)
    cy.get('#selection-icon').should(prefix + 'exist')
}

function clickOnApp() {
    cy.get('#app-list').find('.app-item').click()
}

function clickOnVersion() {
    cy.get('#version-list').find('.version-item').click()
}

function deleteApp() {
    clickOnApp()
    cy.get('#button-delete-app').click();
    cy.get('#button-delete-confirmation').click();
    cy.get('#button-close-popup').should('not.exist');
    cy.get('#button-delete-confirmation').should('not.exist');
}

function tryToDeleteAppButCancelInConfirmationPopup() {
    clickOnApp()
    cy.get('#button-delete-app').click();
    cy.get('#button-close-popup').click();
    cy.get('#button-close-popup').should('not.exist');
    cy.get('#button-delete-confirmation').should('not.exist');
    clickOnApp()
}

function checkInputValidationOnLoginPage() {
    cy.visit(loginPath);
    cy.get('#input-username').type('ad');
    cy.get('#input-password').type('pass');
    shallInvalidInputMessageBeShown(false, "username")
    shallInvalidInputMessageBeShown(false, "password")

    cy.get('#button-login').click();
    shallInvalidInputMessageBeShown(true, "username")
    shallInvalidInputMessageBeShown(true, "password")

    cy.get('#input-username').clear().type(maintainerName);
    cy.get('#input-password').clear().type('password');
    shallInvalidInputMessageBeShown(false, "username")
    shallInvalidInputMessageBeShown(false, "password")
}

function checkInputValidationOnRegistrationPage() {
    cy.visit(registrationPath);
    cy.get('#input-username').type('ad');
    cy.get('#input-password').type('pass');
    cy.get('#input-email').type('a@a');
    shallInvalidInputMessageBeShown(false, "username")
    shallInvalidInputMessageBeShown(false, "password")
    shallInvalidInputMessageBeShown(false, "email")
    cy.get('#terms-and-conditions-acceptance-checkbox').click()

    cy.get('#button-register').click();
    shallInvalidInputMessageBeShown(true, "username")
    shallInvalidInputMessageBeShown(true, "password")
    shallInvalidInputMessageBeShown(true, "email")

    cy.get('#input-username').clear().type(maintainerName);
    cy.get('#input-password').clear().type('password');
    cy.get('#input-email').clear().type('admin@admin.de');
    shallInvalidInputMessageBeShown(false, "username")
    shallInvalidInputMessageBeShown(false, "password")
    shallInvalidInputMessageBeShown(false, "email")
}

function checkInputValidationOnAppPage() {
    cy.get('#go-to-apps').click();
    cy.get('#input-app').type('ad');
    shallInvalidInputMessageBeShown(false, "app")
    cy.get('#button-create-app').click();
    shallInvalidInputMessageBeShown(true, "app")
    cy.get('body').should('contain.text', 'Invalid app,');
    cy.get('#input-app').clear().type('asdf');
    shallInvalidInputMessageBeShown(false, "app")
}

function checkInputValidationOnVersionPage() {
    cy.reload()
    createApp()
    clickOnApp()
    shallInvalidInputMessageBeShown(false, "version")
    cy.get('#button-edit-versions').click()
    cy.get('input[type="file"]').selectFile({
        contents: Cypress.Buffer.from(''),
        fileName: 'as.zip',
    }, {force: true})
    uploadValidVersion()
    shallInvalidInputMessageBeShown(true, "version")
    getVersionItems().click()
    cy.get('#button-delete-version').click()
    cy.get('#button-delete-confirmation').click()
}

function getVersionItems() {
    return cy.get('#version-list').find('.version-item')
}

function logout() {
    cy.get('#logout').click();
    cy.visit(appsPath)
    cy.url().should('eq', loginPath);
    authCookie = ""
}

function uploadValidVersion() {
    cy.task('zipFolderInMemory', "../backend/assets/samplemaintainer-app").then((zipBytes: Buffer) => {
        cy.get('input[type="file"]').selectFile(
            {
                contents: Cypress.Buffer.from(zipBytes),
                fileName: '1.4.zip',
                mimeType: 'application/zip',
            },
            {force: true}
        );
    });
}

function checkInputValidationOnChangePasswordPage() {
    cy.visit(changePasswordPath);
    cy.get('#old_password').type('ad');
    cy.get('#new_password').type('af');
    shallInvalidInputMessageBeShown(false, "password")

    cy.get('#button-change-password').click();
    shallInvalidInputMessageBeShown(true, "password")

    cy.get('#old_password').clear().type('password1')
    cy.get('#new_password').clear().type('password2')
    shallInvalidInputMessageBeShown(false, "password")
}

describe('Hub Operations', () => {

    it('register and login', () => {
        cy.request(hubWipePath)
        cy.visit(appsPath);
        cy.url().should('eq', loginPath);
        cy.get('#go-to-registration-page').click();
        cy.url().should('eq', registrationPath);
        cy.get('#go-to-login-page').click();
        cy.url().should('eq', loginPath);
        cy.get('#go-to-registration-page').click();
        cy.url().should('eq', registrationPath);

        cy.get('#input-username').type(maintainerName);
        cy.get('#input-password').type('password');
        cy.get('#input-email').type('admin@admin.com');
        cy.get('#button-register').should('be.disabled')
        cy.get('#terms-and-conditions-acceptance-checkbox').click()
        cy.get('#button-register').should('be.enabled')
        cy.get('#button-register').click();
        cy.url().should('eq', registrationPath);
        cy.get('body').should('contain.text', 'Your account details have been accepted!');
        cy.get('#go-to-login-page').click();
        cy.url().should('eq', loginPath);
        cy.visit(validationPath + '?code=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef')
        cy.url().should('contain', validationPath);
        cy.get('body').should('contain.text', 'Account validation successful.');
        login()
    });

    it('check input validation', () => {
        checkInputValidationOnLoginPage();
        checkInputValidationOnRegistrationPage();
        login()
        checkInputValidationOnChangePasswordPage();
        checkInputValidationOnAppPage();
        checkInputValidationOnVersionPage();
    });

    it('create and delete app', () => {
        login()
        cy.get('#go-to-apps').click();
        cy.url().should('eq', appsPath);
        createApp()
        assertIsAppSelected(false)
        clickOnApp()
        assertIsAppSelected(true)
        clickOnApp()
        assertIsAppSelected(false)
        tryToDeleteAppButCancelInConfirmationPopup()
        deleteApp()
        assertEmptyAppList();
        cy.get('#go-to-apps').click();
        cy.url().should('eq', appsPath);
    });

    it('check version management page', () => {
        login()
        cy.get('#go-to-apps').click();
        createApp()
        clickOnApp()
        cy.get('#button-edit-versions').click()

        cy.get('#go-to-apps').click();
        cy.url().should('eq', appsPath);
        clickOnApp()
        cy.get('#button-edit-versions').click()

        getVersionItems().should('have.length', 0);
        cy.get('#button-download-version').should('not.exist')
        cy.get('#button-delete-version').should('not.exist')

        cy.url().should('contain', versionManagementPath)

        uploadValidVersion();
        getVersionItems().should('have.length', 1);
        cy.get('#version-name').should('contain', '1.4');
        cy.get('#creation-timestamp').should(($el) => {
            const text = $el.text();
            const regex = /^([A-Za-z]+ \d{1,2}, \d{4}|\d{1,2}\. [A-Za-z]+ \d{4})$/ // Regex for "December 11, 2024" or "3. April 2025"
            expect(text).to.match(regex);
        });

        cy.get('#button-download-version').should('not.exist')
        cy.get('#button-delete-version').should('not.exist')
        cy.get('#button-close-popup').should('not.exist');
        cy.get('#button-delete-confirmation').should('not.exist');

        assertIsVersionSelected(false)
        clickOnVersion()
        assertIsVersionSelected(true)

        cy.get('#button-delete-version').click()
        cy.get('#button-close-popup').click()
        getVersionItems().should('have.length', 1);

        cy.get('#button-delete-version').click()
        cy.get('#button-delete-confirmation').click()
        getVersionItems().should('have.length', 0);

        cy.get('#selected-app').should('contain', appName)
        cy.reload()
        cy.url().should('eq', appsPath);
    });

    it('check logout', () => {
        login()
        logout()
    });

    it('check wrong password prevents login', () => {
        cy.visit(loginPath);
        cy.get('#input-username').type(maintainerName);
        cy.get('#input-password').type('password+x');
        cy.get('#button-login').click();
        cy.url().should('eq', loginPath);
    });

    it('change password', () => {
        login()
        cy.get('#go-to-change-password').invoke('trigger', 'click')
        cy.url().should('eq', changePasswordPath)

        cy.get('#go-to-apps').click()
        cy.url().should('eq', appsPath)
        cy.get('#go-to-change-password').invoke('trigger', 'click')
        cy.url().should('eq', changePasswordPath)

        let newPassword = "password2"
        cy.get('#old_password').type(password)
        cy.get('#new_password').type(newPassword)
        cy.get('#button-change-password').click()
        cy.url().should('eq', appsPath)

        logout()

        cy.get('#input-username').clear().type(maintainerName)
        cy.get('#input-password').clear().type(password)
        cy.get('#button-login').click()
        cy.url().should('eq', loginPath)

        password = newPassword
        login()
    });

    it('page not found', () => {
        login()
        cy.visit(baseUrl + '/not-existing-page')
        cy.get('body').should('contain.text', 'Page Not Found')
        cy.get('#go-to-home').click()
        cy.url().should('eq', appsPath)
    });

    it('test delete account', () => {
        login()
        cancelAccountDeletion();
        executeAccountDeletion();
    });
});

export let password = "password"

function login() {
    if(authCookie == "") {
        cy.visit(loginPath)
        cy.get('#input-username').clear().type(maintainerName)
        cy.get('#input-password').clear().type(password)
        cy.get('#button-login').click()
        cy.url().should('eq', appsPath)
        cy.get('#user-label').should('contain', maintainerName);
        cy.getCookie("auth").should('exist').then((cookie) => {
            authCookie = cookie.value
        })
    } else {
        cy.setCookie("auth", authCookie)
        cy.visit(appsPath)
    }
}

