// This file is not the source of truth for the protocol.
// The protocol is defined by backend/common/proto.go

export interface SocketEvent {
  type: number
  evt: Uint8Array
}

export interface EvtError {
  code: number
}

export interface EvtData {
  guildId: Uint8Array
  evtId: Uint8Array
  timestamp: number
  message: Uint8Array
}

export interface EvtRegister {
  userId: Uint8Array
  token: Uint8Array
}

export interface EvtAuthenticate {
  token: Uint8Array
}

export interface EvtSubGuilds {
  guildIds: Uint8Array[]
}

export interface EvtAddUsers {
  guildId: Uint8Array
  userIds: Uint8Array[]
}

export interface EvtRemoveUsers {
  guildId: Uint8Array
  userIds: Uint8Array[]
}
