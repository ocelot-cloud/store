<template>
  <v-app>
    <v-container max-width="600px" class="mt-5">
      <v-card outlined>
        <v-card-title>
          <span class="text-h5">Registration</span>
        </v-card-title>
        <v-card-text>
          <v-form @submit.prevent="register">
            <ValidatedInput id="input-username" :submitted="submitted" validation-type="username" v-model="user" />
            <ValidatedInput id="input-password" :submitted="submitted" validation-type="password" v-model="password" />
            <ValidatedInput id="input-email" :submitted="submitted" validation-type="email" v-model="email" />

            <v-row align="center">
              <v-col cols="auto">
                <v-checkbox
                    id="terms-and-conditions-acceptance-checkbox"
                    v-model="acceptedTerms"
                    :rules="[v => !!v || 'You must accept the End-User License Agreement']"
                    hide-details
                    density="compact"
                />
              </v-col>
              <v-col>
                <span>
                  I accept the <a href="https://ocelot-cloud.org/docs/legal/eula" target="_blank" rel="noopener noreferrer">End-User License Agreement</a> and agree to be bound by them.
                </span>
              </v-col>
            </v-row>

            <v-btn
                class="mt-4"
                id="button-register"
                type="submit"
                color="primary"
                block
                :disabled="!acceptedTerms"
            >
              Register
            </v-btn>
          </v-form>

          <div id="is-registered-text" v-if="isRegistered" class="mt-4">
            <v-alert type="success" dense>
              Your account details have been accepted! You will shortly receive an email with a verification link to activate your account. Click on the link to complete the process. Once verified, go to the login page and log in with your new account details.
            </v-alert>
          </div>

          <v-btn
              id="go-to-login-page"
              color="secondary"
              block
              class="mt-4"
              @click="router.push(frontendLoginPath)"
          >
            Back to Login Page
          </v-btn>
        </v-card-text>
      </v-card>
    </v-container>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'
import ValidatedInput from '@/components/ValidatedInput.vue'
import { doHubRequest } from '@/components/shared'
import {
  backendAccountRegistrationPath,
  frontendLoginPath,
  frontendTermsPath
} from '@/components/config'

export default defineComponent({
  name: 'HubRegistration',
  components: {  ValidatedInput },
  setup() {
    const router = useRouter()
    const user = ref('')
    const password = ref('')
    const email = ref('')
    const submitted = ref(false)
    const isRegistered = ref(false)
    const showEmailRequirementExplanation = ref(false)
    const acceptedTerms = ref(false)

    const register = async () => {
      submitted.value = true
      const registrationForm = {
        user: user.value,
        password: password.value,
        origin: window.location.origin,
        email: email.value
      }
      const response = await doHubRequest(backendAccountRegistrationPath, registrationForm)
      if (response) {
        isRegistered.value = true
      }
    }

    return {
      user,
      password,
      email,
      register,
      frontendLoginPath,
      frontendTermsPath,
      submitted,
      router,
      isRegistered,
      showEmailRequirementExplanation,
      acceptedTerms,
    }
  }
})
</script>
