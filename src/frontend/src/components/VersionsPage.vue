<template>
  <FrameComponent>
    <h1 class="text-center">Version Management</h1>
    <br>
    <v-row align="center" justify="center">
      <v-col cols="auto" style="max-width: 600px; width: 100%;">
        <v-card outlined>
          <v-card-text>
            <p>Here is a <a href="https://ocelot-cloud.org/docs/app-store/create-own-apps" target="_blank" rel="noopener noreferrer">tutorial</a> helping you get started.</p>
            <br>
            <p id="selected-app">You are currently editing versions of the app <strong>"{{ app }}"</strong>.</p>
            <div class="file-upload-area my-4">
              <input type="file" ref="fileInput" @change="handleFileUpload" class="d-none" />
              <v-sheet
                  id="drag-and-drop-area"
                  elevation="1"
                  rounded
                  class="pa-8 text-center"
                  style="border: 2px dashed #ccc;"
                  @dragover.prevent
                  @drop.prevent="handleDrop"
              >
                <p>Drag and drop the versions zip file here</p>
              </v-sheet>
            </div>

            <v-alert v-if="submitted" type="error" dense>
              {{ errorMessageText }}
            </v-alert>

            <br>

            <h3>Version List</h3>
            <p v-if="!versionList || versionList.length === 0">
              (No versions created yet)
            </p>
            <v-list id="version-list" dense>
              <v-list-item
                  v-for="version in versionList"
                  :key="version.name"
                  :class="{ 'v-item--active': selectedVersion.id === version.id }"
                  @click="selectVersion(version)"
                  class="version-item"
              >
                <div style="display: flex; align-items: center;">
                  <v-list-item-title id="version-name">{{ version.name }}</v-list-item-title>
                  <v-icon id="selection-icon" v-if="selectedVersion.id === version.id" color="success" style="margin-left: 8px;">
                    mdi-check-circle
                  </v-icon>
                </div>

                <v-list-item-subtitle id="creation-timestamp">
                  {{ formatTimestamp(version.creation_timestamp) }}
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>

            <div v-if="versionList && selectedVersion.id" class="d-flex justify-end">
              <v-btn
                  id="button-download-version"
                  color="primary"
                  class="mr-2"
                  @click="downloadVersion"
              >
                Download
              </v-btn>
              <v-btn
                  id="button-delete-version"
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
        :on-confirm="deleteVersion"
        title="Version Deletion Confirmation"
        message="Are you sure you want to delete this version?"
    />
  </FrameComponent>
</template>



<script lang="ts">
import {defineComponent, onMounted, ref} from 'vue';
import { useRouter } from 'vue-router';
import {
  doHubRequest, generateInvalidInputMessage,
  defaultMaxLength,
  defaultMinLength, versionAllowedSymbols,
} from "@/components/shared";
import {alertError} from "@/components/requests";
import DeletionConfirmationDialog from "@/components/DeletionConfirmationDialog.vue";
import {
  backendDeleteVersionPath,
  backendDownloadVersionPath,
  backendGetVersionsPath,
  backendVersionUploadPath,
  frontendAppPath, hubSession
} from "@/components/config";
import FrameComponent from "@/components/FrameComponent.vue";

class Version {
  name: string
  id: string
  creation_timestamp: string

  constructor(name: string, id: string, creationTimeStamp: string) {
    this.name = name;
    this.id = id;
    this.creation_timestamp = creationTimeStamp;
  }
}

export default defineComponent({
  name: "HubVersionManagement",
  components: {FrameComponent, DeletionConfirmationDialog},

  setup() {
    const router = useRouter();
    const versionList = ref<Version[]>([]);
    const selectedVersion = ref<Version>(new Version("", "", ""));
    const showDeleteConfirmation = ref(false);
    const submitted = ref(false);
    const errorMessageText = ref(generateInvalidInputMessage("version", versionAllowedSymbols, defaultMinLength, defaultMaxLength))

    const handleFileUpload = (event: Event) => {
      const files = (event.target as HTMLInputElement).files;
      if (files && files.length > 0) {
        uploadFile(files[0]);
      }
    };

    const uploadFile = (file: File) => {
      const suffix = '.zip';
      if (!file.name.endsWith(suffix)) {
        alert(`The file must have a ${suffix} suffix.`);
        return;
      }

      const version = file.name.slice(0, -suffix.length);

      let regex = new RegExp(`^${versionAllowedSymbols}{${defaultMinLength},${defaultMaxLength}}$`)
      if (!regex.test(version)) {
        submitted.value = true
        return;
      }

      const reader = new FileReader();
      reader.onload = async (event) => {
        const content = btoa(
            String.fromCharCode(...new Uint8Array(event.target?.result as ArrayBuffer))
        );
        const appId = hubSession.selectdAppId
        const versionUpload = {appId, version, content};

        const response = await doHubRequest(backendVersionUploadPath, versionUpload)
        if (response) {
          submitted.value = false
          getVersions()
        }

      };

      reader.onerror = () => {
        console.error('Error reading file');
      };
      reader.readAsArrayBuffer(file);
    };

    const getVersions = async () => {
      const response = await doHubRequest(backendGetVersionsPath, { value: hubSession.selectdAppId });
      if (response != null && response.data) {
        versionList.value = response.data as Version[];
        if (versionList.value.length > 0) {
          versionList.value.sort((a: Version, b: Version) => a.name.localeCompare(b.name));
        }
        for (let version of versionList.value) {
          console.log(version.creation_timestamp)
        }
      }
    };

    const deleteVersion = async () => {
      const response = await doHubRequest(backendDeleteVersionPath, { value: selectedVersion.value.id });
      if (response != null) {
        versionList.value = versionList.value.filter(version => version.id !== selectedVersion.value.id);
        showDeleteConfirmation.value = false;
      } else {
        alert('Failed to delete version.');
      }
    };

    const downloadVersion = async () => {
      try {
        const response = await doHubRequest(backendDownloadVersionPath, { value: selectedVersion.value.id })
        if (response != null) {
          const raw = atob(response.data.content)
          const bytes = new Uint8Array(raw.length)
          for (let i = 0; i < raw.length; i++) {
            bytes[i] = raw.charCodeAt(i)
          }
          const blob = new Blob([bytes], {type: 'application/zip'})
          const url = window.URL.createObjectURL(blob)
          const link = document.createElement('a')
          link.href = url
          link.download = `${selectedVersion.value.name}.zip`
          document.body.appendChild(link)
          link.click()
          link.remove()
          console.log("File download started successfully")
        }
      } catch (error) {
        alertError(error)
        console.error('Error during file download:', error);
      }
    }



    const selectVersion = (version: Version) => {
      if (selectedVersion.value.id == version.id) {
        selectedVersion.value.name = ""
        selectedVersion.value.id = ""
      } else {
        selectedVersion.value.name = version.name;
        selectedVersion.value.id = version.id;
      }
    }

    const handleDrop = (event: DragEvent) => {
      const files = event.dataTransfer?.files;
      if (files && files.length > 0) {
        uploadFile(files[0]);
      }
    };

    const formatTimestamp = (rawTimestamp: string) => {
      const date = new Date(rawTimestamp);
      return new Intl.DateTimeFormat(navigator.language, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      }).format(date);
    };

    onMounted(() => {
      if (hubSession.selectedApp == "") {
        router.push(frontendAppPath)
      }
      getVersions()
    });

    return {
      handleFileUpload,
      versionList,
      app: hubSession.selectedApp,
      selectedVersion,
      deleteVersion,
      selectVersion,
      downloadVersion,
      showDeleteConfirmation,
      handleDrop,
      errorMessageText,
      submitted,
      router,
      formatTimestamp,
      frontendAppPath,
    }
  },
});
</script>
