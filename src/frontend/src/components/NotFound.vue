<template>
  <v-app>
    <v-container class="my-5" justify="center">
      <v-row justify="center">
        <v-col lg="5" md="7" sm="9">
          <v-card class="p-4 bg-dark text-white" elevation="3">
            <v-card-title class="justify-center">
              <h3>Page Not Found</h3>
            </v-card-title>
            <v-card-text>
              <p>The URL path you requested does not exist. Please return to the home page.</p>
              <v-spacer />
              <br>
              <div class="text-center">
                <v-btn id="go-to-home" color="secondary" @click="router.push(frontendAppPath)">
                  Back to Home Page
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
import { defineComponent } from 'vue';
import { useRouter } from 'vue-router';
import {frontendAppPath} from "@/components/config";

export default defineComponent({
  name: 'HubChangePassword',
  setup() {
    const router = useRouter();

    return {
      frontendAppPath,
      router,
    };
  },
});
</script>
