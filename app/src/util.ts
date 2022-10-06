import type { Mutation, Guild } from "./stores/guilds"
import { encode } from "@msgpack/msgpack"
import { parse as uuidParse, v4 as uuidV4 } from "uuid"
import { useEphemeralStore } from "@/stores/ephemeral"
import { useGuildsStore } from "@/stores/guilds"

// For some reason, javascript doesn't have any native mechanism to convert bytes to base64
// or back. I don't know why anyone wants to use this language.
// 
// Maybe I should have just written a native application instead...
export function utob(arr: Uint8Array): string {
  return btoa(
    Array(arr.length)
      .fill('')
      .map((_, i) => String.fromCharCode(arr[i]))
      .join('')
  )
}

export function btou(str: string): Uint8Array {
  return Uint8Array.from(atob(str), (c) => c.charCodeAt(0))
}

// WARNING: deepCopy does not copy functions
export function deepCopy<T>(obj: T): T {
  // Create the (shitty) javascript version of "deep-copy"
  return JSON.parse(JSON.stringify(obj)) as T
}

// This function feels like it could be exploited very easily.
// More research required.
export function applyMut(guild: Guild, mut: Mutation) {
  // Descend tree until we reach object holding desired object (or parent, if it is an array)
  let path = mut.path.split(".").slice(1)
  let obj: any = guild
  for (;;) {
    if (path[0] == "id") {
      throw new Error("cannot update id component of an object")
    } 

    if (path.length == 2 && Array.isArray(obj[path[0]])) {
      break
    } else if (path.length == 1) {
      break
    }

    if (Array.isArray(obj)) {
      obj = obj.filter((v) => v.id == path[0])[0]
    } else {
      obj = obj[path[0]]
    }

    // console.log(path)
    path = path.slice(1)
  }

  // Set desired object
  if (path.length == 2) {
    // Remove item first, then set object if necessary
    const old = obj[path[0]].filter((v: any) => v.id == path[1])[0]
    obj[path[0]] = obj[path[0]].filter((v: any) => v.id != path[1])
    if (mut.method == "SET") {
      if (old.id != mut.object.id) {
        throw new Error("cannot update id component of an object")
      } 

      obj[path[0]].push(mut.object)
    }
  } else {
    if (mut.method == "SET") {
      obj[path[0]] = mut.object
    } else {
      delete obj[path[0]]
    }
  }
}

export function sendMutation(guildId: string, msg: any) {
  const store = useGuildsStore()
  const ephem = useEphemeralStore()

  const evtId = uuidV4()
  ephem.pendingMutations[evtId] = msg

  ephem.ws?.send(
    encode({
      type: 0x03,
      evt: encode({
        guildId: uuidParse(guildId),
        evtId: uuidParse(evtId),
        message: store.txSessions[guildId].sendMessage(
          new TextEncoder().encode(
            JSON.stringify(msg)
          )
        ),
      }),
    })
  )
}
