<script setup lang="ts">
import { useGuildsStore } from "@/stores/guilds"
import ChannelItem from "./ChannelItem.vue"

const guilds = useGuildsStore()
</script>

<template>
  <div class="list">
    <div v-if="guilds.selectedGuild != undefined">
      <ChannelItem
        :text="guilds.selectedGuild.name"
        icon="settings"
      ></ChannelItem>

      <ChannelItem
        v-for="channel in guilds.selectedGuild.channels"
        :key="channel.id"
        :text="channel.name"
        icon="tag"
      ></ChannelItem>
    </div>
  </div>
</template>

<style scoped>
.list {
  background-color: var(--md-surface);
  height: 100%;
  width: 300px;
  border-right: var(--md-surface-variant) 2px solid;
  padding: 10px;
}

.settings p {
  margin: 0;
  padding: 10px;
}
</style>
