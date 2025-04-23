<template>
  <v-app>
    <v-container>
      <v-row align="center" justify="center">
        <v-col cols="12" sm="8" md="5">
          <v-card>
            <v-card-title class="text-h5 text-center">Login</v-card-title>
            <v-card-text>
              <v-form @submit.prevent="login">
                <ValidatedInput
                    id="input-username"
                    validationType="username"
                    v-model="user"
                    :submitted="submitted"
                />
                <ValidatedInput
                    id="input-password"
                    validationType="password"
                    v-model="password"
                    :submitted="submitted"
                />
                <v-btn id="button-login" type="submit" color="primary" block>
                  Login
                </v-btn>
              </v-form>
              <div class="text-center mt-4">
                <v-btn
                    id="go-to-registration-page"
                    color="secondary"
                    @click="router.push(frontendRegistrationPath)"
                    block
                >
                  Go to Registration Page
                </v-btn>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'
import ValidatedInput from '@/components/ValidatedInput.vue'
import { doHubRequest } from '@/components/shared'
import {
  backendLoginPath, frontendAppPath,
  frontendRegistrationPath
} from '@/components/config'

export default defineComponent({
  name: 'HubLogin',
  components: { ValidatedInput },
  setup() {
    const router = useRouter()
    const user = ref('')
    const password = ref('')
    const submitted = ref(false)

    const login = async () => {
      console.log("user: ", user.value)
      console.log("password: ", password.value)
      submitted.value = true
      if (user.value && password.value) {
        const loginForm = {
          user: user.value,
          password: password.value,
          origin: window.origin
        }
        await doHubRequest(backendLoginPath, loginForm)
        router.push(frontendAppPath)
      }
    }

    return {
      user,
      password,
      login,
      frontendRegistrationPath,
      submitted,
      router,
    }
  },
})
</script>
