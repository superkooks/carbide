<script setup lang="ts">
import { useEphemeralStore } from "@/stores/ephemeral"
import { useGlobalStore } from "@/stores/global"
import { onMounted } from "vue"
import { encrypt } from "../util"

const store = useGlobalStore()
const ephem = useEphemeralStore()

function changeName(id: string) {
  ephem.ws?.send(
    encrypt(
      {
        method: "SET",
        path: ".name",
        object: "wat",
      },
      id,
      store.txSessions[id]
    )
  )
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
