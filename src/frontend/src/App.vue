<template>
  <div id="app">
    <router-view></router-view>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import axios from "axios";
import AppsPage from "@/components/AppsPage.vue";

@Options({
  components: {
    AppsPage,
  },
})
export default class App extends Vue {}

axios.defaults.withCredentials = true;
</script>

<style lang="sass">
code
  color: #cf6679 // vuetify's "error" color
</style>
