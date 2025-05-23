import { createRouter, createWebHistory } from 'vue-router';
import axios from "axios";

import {
    frontendAppPath,
    frontendChangePasswordPath, frontendLoginPath,
    frontendRegistrationPath,
    frontendVersionPath, frontendTermsPath, frontendValidationPath,
    hubBaseUrl,
    hubSession,
    Session
} from "@/components/config";
import LoginPage from "@/components/LoginPage.vue";
import RegistrationPage from "@/components/RegistrationPage.vue";
import ChangePassword from "@/components/ChangePassword.vue";
import ChangePasswordPage from "@/components/ChangePasswordPage.vue";
import VersionsPage from "@/components/VersionsPage.vue";
import AppsPage from "@/components/AppsPage.vue";
import ValidationTemplate from "@/components/ValidationTemplate.vue";
import NotFound from "@/components/NotFound.vue";

const routes = [
    {
        path: frontendAppPath,
        name: 'HubAppManagement',
        component: AppsPage,
    },
    {
        path: frontendLoginPath,
        name: 'HubLogin',
        component: LoginPage,
    },
    {
        path: frontendRegistrationPath,
        name: 'HubRegistration',
        component: RegistrationPage,
    },
    {
        path: frontendChangePasswordPath,
        name: 'HubChangePassword',
        component: ChangePasswordPage,
    },
    {
        path: frontendVersionPath,
        name: 'HubVersionManagement',
        component: VersionsPage,
    },
    {
        path: frontendValidationPath,
        name: 'HubValidation',
        component: ValidationTemplate,
    },
    {
        path: '/:pathMatch(.*)*', // Wildcard path to catch all unmatched routes
        name: 'NotFound',
        component: NotFound,
    },
];

const router = createRouter({
    history: createWebHistory(import.meta.env.VITE_BASE_URL),
    routes,
});

async function isThereValidSession(session: Session, apiUrl: string): Promise<boolean> {
    try {
        const response = await axios.get(apiUrl);
        if (response.status === 200) {
            session.user = response.data.value;
            session.isAuthenticated = true;
            return true;
        }
        return false;
    } catch (error) {
        return false;
    }
}

router.beforeEach(async (to, from, next) => {
    if (to.path === frontendLoginPath || to.path === frontendRegistrationPath || to.path === frontendValidationPath || to.path === frontendTermsPath || hubSession.isAuthenticated) {
        next();
    } else {
        if (await isThereValidSession(hubSession, hubBaseUrl + '/api/account/auth-check')) {
            next();
        } else {
            next({ name: 'HubLogin' });
        }
    }
    return;
});

export default router;
