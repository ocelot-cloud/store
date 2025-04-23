
export const hubBaseUrl = import.meta.env.VITE_APP_PROFILE === 'TEST'
    ? 'http://localhost:8082'
    : window.location.origin;

export interface Session {
    user: string;
    isAuthenticated: boolean;
    selectedApp: string,
    selectdAppId: string,
}

export const hubSession: Session = {
    user: "",
    isAuthenticated: false,
    selectedApp: "",
    selectdAppId: "",
};

// FrontendPaths

export const frontendLoginPath = "/login";
export const frontendRegistrationPath = "/registration";
export const frontendAppPath = "/";
export const frontendVersionPath = "/versions";
export const frontendChangePasswordPath = "/change-password";
export const frontendValidationPath = "/validate";
export const frontendTermsPath = "/terms";

// BackendPaths

const accountPrefix = "/account";
export const backendAccountRegistrationPath = accountPrefix + "/registration";
export const backendLoginPath = accountPrefix + "/login";
export const backendChangePasswordPath = accountPrefix + "/change-password";
export const backendAccountLogoutPath = accountPrefix + "/logout";
export const backendAccountDeletePath = accountPrefix + "/delete";
export const backendValidationPath = accountPrefix + "/validate";

const appsPrefix = "/apps";
export const backendAppCreationPath = appsPrefix + "/create";
export const backendGetListsPath = appsPrefix + "/get-list";
export const backendDeleteAppsPath =  appsPrefix + "/delete";

const versionsPrefix = "/versions";
export const backendGetVersionsPath = versionsPrefix + "/list";
export const backendVersionUploadPath = versionsPrefix + "/upload";
export const backendDownloadVersionPath = versionsPrefix + "/download";
export const backendDeleteVersionPath = versionsPrefix + "/delete";
