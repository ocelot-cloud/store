<template>
  <FrameComponent>
    <h1 class="text-center">Change Password</h1>
    <br>
    <v-row align="center" justify="center">
      <v-col cols="auto" style="max-width: 400px; width: 100%;">
        <v-card outlined>
          <v-card-text>
            <v-form @submit.prevent="changePassword">
              <ValidatedInput
                  id="old_password"
                  :submitted="submitted"
                  validation-type="password"
                  v-model="oldPassword"
                  label="Enter Old Password"
              />
              <ValidatedInput
                  id="new_password"
                  :submitted="submitted"
                  validation-type="password"
                  v-model="newPassword"
                  label="Enter New Password"
              />
              <v-btn id="button-change-password" type="submit" color="primary" block>
                Change Password
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </FrameComponent>
</template>


<script lang="ts">
import { defineComponent, ref } from 'vue';
import {doHubRequest} from "@/components/shared";
import { useRouter } from 'vue-router';
import {backendChangePasswordPath} from "@/components/config";
import FrameComponent from "@/components/FrameComponent.vue";
import ValidatedInput from "@/components/ValidatedInput.vue";

export default defineComponent({
  name: 'HubChangePassword',
  components: {ValidatedInput, FrameComponent},
  setup() {
    const router = useRouter();
    const user = ref('');
    const oldPassword = ref('');
    const newPassword = ref('');
    const submitted = ref(false);

    const changePassword = async () => {
      submitted.value = true
      const changePasswordForm = { old_password: oldPassword.value, new_password: newPassword.value };
      let response = await doHubRequest(backendChangePasswordPath, changePasswordForm)
      if (response && response.status === 200) {
        router.push("/")
      }
    };

    const redirectToHubHomePage = () => {
      router.push("/")
    }

    return {
      user,
      oldPassword,
      newPassword,
      changePassword,
      redirectToHubHomePage,
      router,
      submitted,
    };
  },
});
</script>
