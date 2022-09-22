<script setup lang="ts">
import Guilds from "./components/GuildList.vue";
import Channels from "./components/ChannelList.vue";
import MainView from "./components/MainView.vue";
import { useGlobalStore } from "@/stores/global";
import { Go } from "@/assets/wasm_exec.js";
import { set as idbset, get as idbget } from "idb-keyval";
import { parse as uuidparse, v4 as uuidv4 } from "uuid";

const store = useGlobalStore();

// Load the state from store
idbget("state").then((v) => {
  if (v == "") {
    return
  }

  store.overwrite(JSON.parse(v as string));
});

// Listen for updates
store.$subscribe((mutation, state) => {
  console.log(JSON.stringify(state));
  idbset("state", JSON.stringify(state));
});

// Fetch and start tungsten
const go = new Go();
WebAssembly.instantiateStreaming(
  fetch("/tungsten.wasm"),
  go.importObject
).then((result) => {
  go.run(result.instance);
  const txs = window.tungsten.doubleTx();
  const alice = txs[0];
  const bob = txs[1];

  const a2 = window.tungsten.importTx(alice.export());

  const aup = a2.generateUpdate();
  bob.receiveMessage(aup);

  const bup = bob.generateUpdate();
  a2.receiveMessage(bup);

  const ctext = a2.sendMessage(new TextEncoder().encode("testing 123"));
  console.log(ctext);
  console.log(new TextDecoder().decode(bob.receiveMessage(ctext)));

  // Open websocket connection to backend
  const socket = new WebSocket("ws://" + window.location.hostname + ":8080/ws")
  socket.onopen = () => {
    console.log("hey~!")

    const guilds = ["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b", uuidv4()]
    console.log("guilds", guilds)
    socket.send(window.tungsten.helpers.marshalSubGuilds(guilds))

    setTimeout(() => {
      socket.send(window.tungsten.helpers.marshalData("6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b", new TextEncoder().encode("lololol")))
    }, 1000)

    // socket.send()
  }
  socket.onmessage = () => {
    console.log("msg")
  }

});
</script>

<template>
  <!-- Unfortunate inline styling due to globablly scoped css -->
  <div style="display: flex; flex-direction: row; height: 100%">
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
