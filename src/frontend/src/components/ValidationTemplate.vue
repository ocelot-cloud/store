<template>
  <div class="container my-5">
    <div class="row justify-content-center">
      <div class="col-lg-5 col-md-7 col-sm-9">
        <div class="entity-management-container p-4 shadow-sm bg-dark rounded">
          <h3 class="text-center mb-4">Validation</h3>
          <div v-if="message" :class="messageClass">{{ message }}</div>
          <p class="text-center mt-3">
            Back to <a @click.prevent="redirectToLogin" href="#" class="text-primary">login</a>.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {backendValidationPath, frontendLoginPath} from "@/components/config";
import {doHubRequest} from "@/components/shared";

export default defineComponent({
  name: 'HubChangePassword',
  setup() {
    const router = useRouter();
    const route = useRoute();
    const message = ref<string | null>(null);
    const messageClass = ref<string>('text-center text-success');

    onMounted(async () => {
      const code = route.query.code;
      if (!code) {
        message.value = 'Error: Validation code missing.';
        messageClass.value = 'text-center text-danger';
        return;
      }

      let response = await doHubRequest(backendValidationPath + '?code=' + code, null)
      if (response != null && response.status === 200) {
        message.value = 'Account validation successful.';
      } else {
        message.value = 'Error: Account validation failed.';
        messageClass.value = 'text-center text-danger';
      }
    });

    const redirectToLogin = () => {
      router.push(frontendLoginPath)
    };

    return {
      message,
      messageClass,
      redirectToLogin,
    };
  },
});
</script>
