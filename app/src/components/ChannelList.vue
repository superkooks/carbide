<script setup lang="ts">
import { reactive } from "vue"
import ChannelItem from "./ChannelItem.vue"

const state = reactive({
  channels: [
    {
      id: "text",
    },
  ],
})
</script>

<template>
  <div class="list">
    <ChannelItem
      v-for="channel in state.channels"
      :key="channel.id"
      :text="channel.id"
    ></ChannelItem>
  </div>
</template>

<style scoped>
.list {
  background-color: var(--md-surface);
  height: 100%;
  width: 300px;
}
</style>
