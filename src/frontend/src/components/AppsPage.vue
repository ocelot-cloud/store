<template>
  <FrameComponent>
    <h1 class="text-center">My Apps</h1>
    <br>
    <v-row align="center" justify="center">
      <v-col cols="auto" style="max-width: 600px; width: 100%;">
        <v-card outlined>
          <v-card-text>
            <p>Tip: Apps typically have the name of the software they are meant to deploy, such as 'gitlab', 'wordpress', etc.</p>
            <br>
            <v-row>
              <v-col style="max-width: 260px; width: 100%;">
                <ValidatedInput
                    id="input-app"
                    :submitted="submitted"
                    validation-type="app"
                    v-model="newAppToCreate"
                />
              </v-col>
              <v-col style="max-width: 160px; width: 100%;">
                <v-btn id="button-create-app" color="primary" block @click="createApp" style="margin-top: 10px">
                  Create App
                </v-btn>
              </v-col>
            </v-row>

            <v-divider class="my-4"></v-divider>

            <div>
              <h3>App List</h3>
              <p v-if="!appList || appList.length === 0">(No apps created yet)</p>
              <v-list id="app-list" dense>
                <v-list-item
                    class="app-item"
                    v-for="app in appList"
                    :key="app.name"
                    :class="{ 'v-item--active': selectedApp.id === app.id }"
                    @click="selectApp(app)"
                >
                  <div style="display: flex; align-items: center;">
                    <v-list-item-title>{{ app.name }}</v-list-item-title>
                    <v-icon id="selection-icon" v-if="selectedApp.id === app.id" color="success" style="margin-left: 8px;">
                      mdi-check-circle
                    </v-icon>
                  </div>
                </v-list-item>
              </v-list>
            </div>

            <v-divider class="my-4"></v-divider>

            <div v-if="appList && selectedApp.id != ''" class="d-flex justify-end">
              <v-btn
                  id="button-edit-versions"
                  color="primary"
                  class="mr-2"
                  @click="goToVersionManagement()"
              >
                Edit Versions
              </v-btn>
              <v-btn
                  id="button-delete-app"
                  color="error"
                  @click="showDeleteConfirmation = true"
              >
                Delete
              </v-btn>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <DeletionConfirmationDialog
        v-model:visible="showDeleteConfirmation"
        :on-confirm="deleteApp"
        title="App Deletion Confirmation"
        message="Are you sure you want to delete this app? This will also delete all associated versions."
    />
  </FrameComponent>
</template>

<script lang="ts">
import {defineComponent, onMounted, ref} from "vue";
import { useRouter } from 'vue-router';
import DeletionConfirmationDialog from "@/components/DeletionConfirmationDialog.vue";
import ValidatedInput from "@/components/ValidatedInput.vue";
import {doHubRequest} from "@/components/shared";
import {
  backendAppCreationPath,
  backendDeleteAppsPath,
  backendGetListsPath,
  frontendVersionPath,
  hubSession
} from "@/components/config";
import FrameComponent from "@/components/FrameComponent.vue";

class App {
  name: string;
  id: string;

  constructor(name: string, id: string) {
    this.name = name;
    this.id = id;
  }
}

export default defineComponent({
  name: 'HubAppManagement',
  components: {FrameComponent, ValidatedInput, DeletionConfirmationDialog},

  setup() {
    const router = useRouter();
    const user = ref("");
    const showDeleteConfirmation = ref(false);
    const newAppToCreate = ref('');
    const appList = ref<App[]>([]);
    const selectedApp = ref<App>(new App("", ""));
    const isEditingVersions = ref(false);
    const submitted = ref(false);

    const selectApp = (app: App) => {
      if (selectedApp.value.id == app.id) {
        selectedApp.value.id = ""
      } else {
        selectedApp.value.id = app.id;
        selectedApp.value.name = app.name;
      }
    };

    const goToVersionManagement = () => {
      hubSession.user = user.value
      hubSession.selectedApp = selectedApp.value.name
      hubSession.selectdAppId = selectedApp.value.id
      router.push(frontendVersionPath)
    }

    const createApp = async () => {
      submitted.value = true
      const response = await doHubRequest(backendAppCreationPath, { value: newAppToCreate.value })
      if (response) {
        await getApps()
        newAppToCreate.value = ""
        submitted.value = false
      }
    };

    const getApps = async () => {
      const response = await doHubRequest(backendGetListsPath, null)
      if (response != null && response.data) {
        appList.value = response.data as App[];
        if (appList.value.length > 0) {
          appList.value.sort((a: App, b: App) => a.name.localeCompare(b.name));
        }
      }
    };

    const deleteApp = async () => {
      const response = await doHubRequest(backendDeleteAppsPath, { value: selectedApp.value.id })
      if (response != null) {
        await getApps()
        appList.value = appList.value.filter(app => app.id !== selectedApp.value.id);
        selectedApp.value.id = ""
        showDeleteConfirmation.value = false
      }
    };

    onMounted(() => {
      user.value = hubSession.user
      getApps();
    });

    return {
      isEditingVersions,
      appList,
      selectedApp,
      newAppToCreate,
      selectApp,
      goToVersionManagement,
      createApp,
      deleteApp,
      showDeleteConfirmation,
      submitted,
      router,
    }
  },
})
</script>