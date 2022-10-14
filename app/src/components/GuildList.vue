<script setup lang="ts">
import { useEphemeralStore } from "@/stores/ephemeral"
import { useGuildsStore } from "@/stores/guilds"
import { sendMutation, ZERO_UUID } from "../util"
import { encode } from "@msgpack/msgpack"
import { parse, v4 } from "uuid"
import { useUserStore } from "@/stores/user"

const guilds = useGuildsStore()
const ephem = useEphemeralStore()
const user = useUserStore()

function createGuild() {
  // Create guild id
  const newGuild = v4()

  // Generate new tx
  guilds.txSessions[newGuild] = window.tungsten.genTx(user.id)

  // Sub to new guild
  ephem.ws?.send(
    encode({
      type: 0x06,
      evt: encode({
        guildIds: [parse(newGuild)],
      }),
    })
  )

  // Send initial mutation
  sendMutation(newGuild, ZERO_UUID, {
    method: "SET",
    path: ".",
    object: {
      id: newGuild,
      name: "new guild",
      channels: [],
    },
  })
}
</script>

<template>
  <div class="container">
    <div v-for="guild in guilds.guilds" :key="guild.id">
      <div class="guild">
        <!-- <span class="selectedIndicator" v-show="selected"></span> -->
        <!-- <span class="unreadIndicator" v-show="unread"></span> -->
        <!-- <img
          class="guildIcon"
          v-if="guild.icon != ''"
          :src="$guilds.state.hostedDomain + guild.icon"
          @click="select"
        /> -->
        <p class="guildIcon text">
          {{ guild.name.substring(0, 1) }}
        </p>

        <div class="name title-medium">
          <p>{{ guild.name }}</p>
        </div>
      </div>
    </div>

    <div class="guild">
      <p class="guildIcon text" style="color: #3ba55d" @click="createGuild">
        +
      </p>
      <div class="name title-medium">
        <p>Create a guild</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  background-color: var(--md-surface);
  height: calc(100% - 20px);
  padding: 10px;

  border-right: var(--md-surface-variant) 2px solid;
}

.guild {
  display: flex;
  flex-direction: row;
  position: relative;
}

/* Guild name tag */
.guild .name {
  position: absolute;
  left: 65px;
  top: 3.5px;
  padding: 7px;
  padding-left: 10px;
  padding-right: 10px;

  background-color: var(--md-on-secondary);
  border-radius: 5px;
  opacity: 0%;
  transition: opacity 0.1s ease-out;
  pointer-events: none;

  z-index: 1;
}

.guild .name p {
  margin: 0;
  width: max-content;
}

.guildIcon:hover ~ .name {
  opacity: 100%;
}

/* Guild icon */
.guildIcon {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  cursor: pointer;
  margin-bottom: 8px;

  transition: border-radius 0.3s;
}

.guildIcon.text {
  height: 40px;
  padding-top: 5px;
  font-family: "Roboto", sans-serif;
  font-size: 30px;
  text-align: center;
  margin-top: 0px;
  background: var(--md-on-secondary);
}

.guildIcon:hover {
  border-radius: 30%;
}

/* Markers */
.selected .guildIcon {
  border-radius: 30%;
  margin-left: 7px;
}

.unread .guildIcon {
  margin-left: 7px;
}

.selectedIndicator {
  border-top-right-radius: 5px;
  border-bottom-right-radius: 5px;
  padding-right: 4px;
  height: 35px;
  margin-top: 9.5px;
  background-color: #ffffff;
}

.unreadIndicator {
  border-top-right-radius: 5px;
  border-bottom-right-radius: 5px;
  padding-right: 4px;
  height: 10px;
  margin-top: 22px;
  background-color: #ffffff;
}
</style>
