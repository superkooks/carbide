<script setup lang="ts">
import Guilds from "./components/GuildList.vue";
import Channels from "./components/ChannelList.vue";
import IndexedStore from "./components/IndexedStore.vue";
import MainView from "./components/MainView.vue";
</script>

<template>
  <!-- Unfortunate inline styling due to globablly scoped css -->
  <div style="display: flex; flex-direction: row; height: 100%">
    <IndexedStore></IndexedStore>

    <Guilds></Guilds>
    <Channels></Channels>
    <MainView></MainView>
  </div>
</template>

<style>
@import "./assets/theme.css";
@import url("https://fonts.googleapis.com/css2?family=Roboto:wght@400;700&display=swap");
@import url("https://fonts.googleapis.com/css2?family=Material+Symbols+Rounded:opsz,wght,FILL,GRAD@48,400,1,0");

html,
#app {
  height: 100%;
}

body {
  color: var(--md-on-background);
  background: var(--md-surface1);
  height: 100%;
  margin: 0;
}
</style>
