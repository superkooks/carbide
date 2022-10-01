<script setup lang="ts">
import Guilds from "./components/GuildList.vue"
import Channels from "./components/ChannelList.vue"
import MainView from "./components/MainView.vue"
import { useGlobalStore, type Guild, type Mutation } from "@/stores/global"
import { Go } from "@/assets/wasm_exec.js"
import { set as idbset, get as idbget } from "idb-keyval"
import { utob, btou, applyMut } from "./util"
import { useEphemeralStore } from "./stores/ephemeral"

const store = useGlobalStore()
const ephem = useEphemeralStore()

// Fetch and start tungsten
const go = new Go()
WebAssembly.instantiateStreaming(fetch("/tungsten.wasm"), go.importObject).then(
  (result) => {
    go.run(result.instance)

    // Load the state from store
    idbget("state").then((v) => {
      if (v == "" || v == undefined) {
        return
      }

      store.overwrite(JSON.parse(v as string))
    })

    // Listen for updates
    store.$subscribe((mutation, state) => {
      const e = store.export()
      // console.log(e)
      idbset("state", e)
    })

    // Open websocket connection to backend
    const socket = new WebSocket(
      "ws://" + window.location.hostname + ":8080/ws"
    )
    socket.onopen = () => {
      ephem.ws = socket

      // On open, subscribe to guilds
      const guilds = ["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"]
      socket.send(window.tungsten.helpers.marshalSubGuilds(guilds))
    }

    socket.onmessage = (v) => {
      ;(v.data as Blob).arrayBuffer().then((ab) => {
        const evt = new Uint8Array(ab)

        const evtType = window.tungsten.helpers.eventType(evt)
        console.log("new msg type:", evtType)
        if (evtType == "DATA") {
          const { guild, msg } = window.tungsten.helpers.unmarshalData(evt)

          console.log("msg guild:", guild)

          const txt = new TextDecoder().decode(
            store.txSessions[guild].receiveMessage(msg)
          )
          console.log(txt)
          const mut = JSON.parse(txt) as Mutation

          applyMut(store.guilds[guild], mut)
        }
      })
    }

    window.genTx = function () {
      const txs = window.tungsten.doubleTx()

      store.txSessions["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] = txs[0]
      return utob(txs[1].export())
    }

    window.setTx = function (input: string) {
      store.txSessions["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] =
        window.tungsten.importTx(btou(input))
    }
  }
)
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
