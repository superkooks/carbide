<script setup lang="ts">
import Guilds from "./components/GuildList.vue";
import Channels from "./components/ChannelList.vue";
import MainView from "./components/MainView.vue";
import { useGlobalStore, type Guild } from "@/stores/global";
import { Go } from "@/assets/wasm_exec.js";
import { set as idbset, get as idbget } from "idb-keyval";
import { utob, btou, applyMut } from "./util";

const store = useGlobalStore();

let g: Guild = {
  id: "test",
  name: "ahh",
  channels: [{
    id: "sub",
    name: "test"
  }]
}
console.log(JSON.stringify(g))
applyMut(g, {
  method: "DELETE",
  path: ".channels.sub.id",
})
console.log(JSON.stringify(g))

// Fetch and start tungsten
const go = new Go();
WebAssembly.instantiateStreaming(
  fetch("/tungsten.wasm"),
  go.importObject
).then((result) => {
  go.run(result.instance);

  // Load the state from store
  idbget("state").then((v) => {
    if (v == "" || v == undefined) {
      return
    }

    store.overwrite(JSON.parse(v as string));
  });

  // Listen for updates
  store.$subscribe((mutation, state) => {
    let e = store.export()
    // console.log(e)
    idbset("state", e);
  });

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
      if (evtType == "DATA") {
        let { guild, msg } = window.tungsten.helpers.unmarshalData(evt)

        console.log("msg guild:", guild)

        console.log(new TextDecoder().decode(window.ourTx.receiveMessage(msg)))
      }
    })
  }

  window.genTx = function () {
    const txs = window.tungsten.doubleTx();

    store.txSessions["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] = txs[0]
    return utob(txs[1].export())
  }

  window.setTx = function (input: string) {
    store.txSessions["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] = window.tungsten.importTx(btou(input))
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
