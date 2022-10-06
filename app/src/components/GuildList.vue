<script setup lang="ts">
import { useEphemeralStore } from "@/stores/ephemeral"
import { useGuildsStore } from "@/stores/guilds"
import { sendMutation } from "../util"

const store = useGuildsStore()
const ephem = useEphemeralStore()

function changeName(id: string) {
  sendMutation(id, {
    method: "SET",
    path: ".name",
    object: "wat",
  })
}
</script>

<template>
  <div class="container">
    <div
      v-for="guild in store.guilds"
      :key="guild.id"
      @click="changeName(guild.id)"
    >
      {{ guild.name }}
    </div>
  </div>
</template>

<style scoped>
.container {
  background-color: var(--md-surfaceneg1);
  height: 100%;
  width: 70px;
}
</style>
