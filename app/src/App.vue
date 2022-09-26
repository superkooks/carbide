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
  if (v == "" || v == undefined) {
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

  // Open websocket connection to backend
  const socket = new WebSocket("ws://" + window.location.hostname + ":8080/ws")
  socket.onopen = () => {
    // On open, subscribe to guilds
    const guilds = ["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"]
    socket.send(window.tungsten.helpers.marshalSubGuilds(guilds))
  }

  socket.onmessage = (v) => {
    (v.data as Blob).arrayBuffer().then(ab => {
      let evt = new Uint8Array(ab)

      let evtType = window.tungsten.helpers.eventType(evt)
      console.log("new msg type:", evtType)
      switch (evtType) {
        case "DATA":
          let { guild, msg } = window.tungsten.helpers.unmarshalData(evt)

          console.log("msg guild:", guild)

          console.log(new TextDecoder().decode(window.ourTx.receiveMessage(msg)))
      }
    })
  }

  let utob = (arr: Uint8Array): string =>
    btoa(
      Array(arr.length)
        .fill('')
        .map((_, i) => String.fromCharCode(arr[i]))
        .join('')
    );

  const btou = (str: string): Uint8Array => Uint8Array.from(atob(str), (c) => c.charCodeAt(0));

  // let btou = (str: string): Uint8Array =>
  //   atob(str) 

  window.genTx = function () {
    const txs = window.tungsten.doubleTx();

    window.ourTx = txs[0]
    return utob(txs[1].export())
  }

  window.setTx = function (input: string) {
    window.ourTx = window.tungsten.importTx(btou(input))
  }

  window.sendUpdate = function () {
    socket.send(window.tungsten.helpers.marshalData("6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b", window.ourTx.generateUpdate()))
  }

  window.sendMsg = function () {
    socket.send(window.tungsten.helpers.marshalData("6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b", window.ourTx.sendMessage(new TextEncoder().encode("fuck yeha"))))
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
