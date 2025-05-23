<template>
  <v-app>
    <v-app-bar flat>
      <v-row align="center" justify="space-between" class="pa-2">
        <v-col cols="auto">
          <img src="../assets/logo.png" alt="Logo" style="height: 45px;" />
        </v-col>
        <v-col class="white--text text-h5">
          Ocelot App Store
        </v-col>
      </v-row>
    </v-app-bar>
    <v-navigation-drawer app v-model="drawer" permanent>
      <v-toolbar flat>
        <v-toolbar-title>Navigation</v-toolbar-title>
      </v-toolbar>
      <v-divider></v-divider>
      <v-list>
        <v-list-item id="go-to-apps" @click="router.push(frontendAppPath)" prepend-icon="mdi-apps">
          <v-list-item-title>My Apps</v-list-item-title>
        </v-list-item>
        <v-list-item id="go-to-change-password" @click="router.push(frontendChangePasswordPath)" prepend-icon="mdi-lock-reset">
          <v-list-item-title>Change Password</v-list-item-title>
        </v-list-item>
        <v-list-item id="logout" @click="logout" prepend-icon="mdi-logout">
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
        <v-list-item id="delete-account" @click="showDeleteConfirmation = true" prepend-icon="mdi-account-remove">
          <v-list-item-title>Delete Account</v-list-item-title>
        </v-list-item>
      </v-list>
      <v-divider></v-divider>
      <div id="user-label" class="pa-4 text-center">
        Logged in as: <strong>{{ hubSession.user }}</strong>
      </div>
    </v-navigation-drawer>

    <v-main>
      <v-container>
        <slot />
      </v-container>
    </v-main>

    <DeletionConfirmationDialog
        v-model:visible="showDeleteConfirmation"
        :on-confirm="deleteAccount"
        title="Account Deletion Confirmation"
        message="Are you sure you want to delete your account?"
    />
  </v-app>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  backendAccountDeletePath,
  backendAccountLogoutPath,
  frontendAppPath, frontendChangePasswordPath,
   frontendLoginPath, hubSession,
} from '@/components/config'
import {doHubRequest} from "@/components/shared";
import DeletionConfirmationDialog from "@/components/DeletionConfirmationDialog.vue";

export default defineComponent({
  name: 'frame-component',
  components: {DeletionConfirmationDialog},

  setup() {
    const drawer = ref(true)
    const router = useRouter()

    const showDeleteConfirmation = ref(false)

    const logout = async () => {
      await doHubRequest(backendAccountLogoutPath, null)
      hubSession.user = ''
      hubSession.isAuthenticated = false
      router.push(frontendLoginPath)
    }

    const deleteAccount = async () => {
      await doHubRequest(backendAccountDeletePath, null)
      hubSession.user = ''
      hubSession.isAuthenticated = false
      router.push(frontendLoginPath)
    }

    return { drawer, router, hubSession, frontendAppPath, frontendChangePasswordPath, logout, deleteAccount, showDeleteConfirmation }
  }
})
</script>

<style scoped lang="sass">
.v-navigation-drawer
  width: 256px
</style>