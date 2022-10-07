<script setup lang="ts">
import GuildList from "./components/GuildList.vue"
import ChannelList from "./components/ChannelList.vue"
import MainView from "./components/MainView.vue"
import { useGuildsStore, type Mutation } from "@/stores/guilds"
import { Go } from "@/assets/wasm_exec.js"
import { set as idbset, get as idbget } from "idb-keyval"
import { utob, btou, applyMut } from "./util"
import { useEphemeralStore } from "./stores/ephemeral"
import { useUserStore } from "./stores/user"
import { encode, decode } from "@msgpack/msgpack"
import { stringify as uuidStringify, parse as uuidParse } from "uuid"
import type { SocketEvent, EvtData, EvtError, EvtRegister } from "./wsevent"

const guilds = useGuildsStore()
const user = useUserStore()
const ephem = useEphemeralStore()

// Fetch and start tungsten
const go = new Go()
WebAssembly.instantiateStreaming(fetch("/tungsten.wasm"), go.importObject).then(
  (result) => {
    go.run(result.instance)

    // Load the state from guilds
    idbget("guilds").then((v) => {
      if (v == "" || v == undefined) {
        return
      }

      guilds.overwrite(JSON.parse(v as string))
    })
    idbget("user").then((v) => {
      if (v == "" || v == undefined) {
        return
      }

      user.overwrite(JSON.parse(v as string))
    })

    // Listen for updates
    guilds.$subscribe((mutation, state) => {
      const e = guilds.export()
      // console.log(e)
      idbset("guilds", e)
    })
    user.$subscribe((mutation, state) => {
      const e = user.export()
      idbset("user", e)
    })

    // Open websocket connection to backend
    const socket = new WebSocket(
      "ws://" + window.location.hostname + ":8080/ws"
    )

    socket.onopen = () => {
      ephem.ws = socket

      // On open, authenticate
      if (user.token == null) {
        // We don't have a token, we should register instead
        // TODO Eventually I will get around to actually adding a register screen
        socket.send(
          encode({
            type: 0x04, // TODO use constants
          })
        )
      } else {
        socket.send(
          encode({
            type: 0x05,
            evt: encode({
              token: user.token,
            }),
          })
        )
      }

      socket.send(
        encode({
          type: 0x06,
          evt: encode({
            guildIds: Object.keys(guilds.guilds).map((v) => uuidParse(v)),
          }),
        })
      )
    }

    socket.onmessage = (v) => {
      ;(v.data as Blob).arrayBuffer().then((ab) => {
        const event = decode(new Uint8Array(ab)) as SocketEvent

        console.log("new msg type:", event.type)
        if (event.type == 0x00) {
          // Respond with heartbeat ack
          socket.send(
            encode({
              type: 0x01,
            })
          )
        } else if (event.type == 0x02) {
          const { code } = decode(event.evt) as EvtError
          console.log("websocket non-fatal error:", code)
        } else if (event.type == 0x03) {
          const { guildId, message, timestamp, evtId } = decode(
            event.evt
          ) as EvtData
          const guild = uuidStringify(guildId)

          console.log("msg guild:", uuidStringify(guildId))

          const { msg, error } =
            guilds.txSessions[guild].receiveMessage(message)

          if (!error) {
            const txt = new TextDecoder().decode(msg)
            console.log(txt)
            const mut = JSON.parse(txt) as Mutation
            guilds.latestTs[guild] = timestamp

            if (
              guilds.guilds[guild] == undefined &&
              mut.path == "." &&
              mut.method == "SET"
            ) {
              guilds.guilds[guild] = mut.object
            } else {
              applyMut(guilds.guilds[guild], mut)
            }
          } else {
            // If there is an error, then it is probably because we sent the message,
            // so we should look for any pending mutations
            const mut = ephem.pendingMutations[uuidStringify(evtId)]
            if (mut != undefined) {
              guilds.latestTs[guild] = timestamp

              if (guilds.guilds[guild] == undefined && mut.path == ".") {
                guilds.guilds[guild] = mut.object
              } else {
                applyMut(guilds.guilds[guild], mut)
              }
            }
          }
        } else if (event.type == 0x04) {
          const { userId, token } = decode(event.evt) as EvtRegister
          user.id = uuidStringify(userId)
          user.token = token
        }
      })
    }

    socket.onclose = (v) => {
      console.log("websocket fatal error:", v.code)
    }

    window.genTx = function () {
      const txs = window.tungsten.doubleTx()

      guilds.txSessions["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] = txs[0]
      return utob(txs[1].export())
    }

    window.setTx = function (input: string) {
      guilds.txSessions["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] =
        window.tungsten.importTx(btou(input))
    }
  }
)
</script>

<template>
  <!-- Unfortunate inline styling due to globablly scoped css -->
  <div style="display: flex; flex-direction: row; height: 100%">
    <GuildList></GuildList>
    <ChannelList></ChannelList>
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
  background: var(--md-background);
  height: 100%;
  margin: 0;
}
</style>
