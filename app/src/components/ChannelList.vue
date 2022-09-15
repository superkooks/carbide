<script setup lang="ts">
import { reactive } from "vue";
import Channel from "./Channel.vue";

type Channell = {
  id: string
}

const state = reactive({
  channels: [
    {
      id: "text",
    },
  ],
});
</script>

<template>
  <div class="list">
    <Channel v-for="channel in state.channels" :key="channel.id" :text="channel.id"></Channel>
  </div>
</template>

<style scoped>
.list {
  background-color: var(--md-surface);
  height: 100%;
  width: 300px;
}
</style>
