package main

import (
	"bytes"
	"syscall/js"
)

// The signatures of all of these functions are available under
// tungsten in /app/src/assets/wasm_exec.ts

// This file is my least favourite.

func genGlobalJS() {
	obj := js.ValueOf(map[string]interface{}{})

	obj.Set("helpers", genHelpers())
	obj.Set("genTx", js.FuncOf(genTxWrapped))
	obj.Set("importTx", js.FuncOf(importTxWrapped))

	// TODO: Remove temp functions
	obj.Set("doubleTx", js.FuncOf(doubleTxWrapped))

	js.Global().Set("tungsten", obj)
}

func genTxWrapped(this js.Value, args []js.Value) any {
	return populateTxMethods(GenTx())
}

func importTxWrapped(this js.Value, args []js.Value) any {
	buf := make([]byte, args[0].Length())
	js.CopyBytesToGo(buf, args[0])

	return populateTxMethods(ImportTx(bytes.NewBuffer(buf)))
}

// TODO: Remove temp function
func doubleTxWrapped(this js.Value, args []js.Value) any {
	alice := GenTx()
	bob := GenTx()

	RxFromTx(bob, alice)
	RxFromTx(alice, bob)

	arr := js.ValueOf([]interface{}{})
	arr.SetIndex(0, populateTxMethods(alice))
	arr.SetIndex(1, populateTxMethods(bob))

	return arr
}

func populateTxMethods(tx *TxSession) js.Value {
	send := func(this js.Value, args []js.Value) any {
		msg := make([]byte, args[0].Length())
		js.CopyBytesToGo(msg, args[0])

		b := new(bytes.Buffer)
		tx.SendMessage(msg, b)

		out := js.Global().Get("Uint8Array").New(b.Len())
		js.CopyBytesToJS(out, b.Bytes())
		return out
	}

	receive := func(this js.Value, args []js.Value) any {
		in := make([]byte, args[0].Length())
		js.CopyBytesToGo(in, args[0])

		msg := tx.ReceiveMessage(in)
		out := js.Global().Get("Uint8Array").New(len(msg))
		js.CopyBytesToJS(out, msg)
		return out
	}

	genUpdate := func(this js.Value, args []js.Value) any {
		b := new(bytes.Buffer)
		tx.GenerateUpdate(b)

		out := js.Global().Get("Uint8Array").New(b.Len())
		js.CopyBytesToJS(out, b.Bytes())
		return out
	}

	export := func(this js.Value, args []js.Value) any {
		b := new(bytes.Buffer)
		tx.Export(b)

		out := js.Global().Get("Uint8Array").New(b.Len())
		js.CopyBytesToJS(out, b.Bytes())
		return out
	}

	obj := js.ValueOf(map[string]interface{}{
		"sendMessage":    js.FuncOf(send),
		"receiveMessage": js.FuncOf(receive),
		"generateUpdate": js.FuncOf(genUpdate),
		"export":         js.FuncOf(export),
	})

	return obj
}

func genHelpers() js.Value {
	eventType := func(this js.Value, args []js.Value) any {
		evt := make([]byte, args[0].Length())
		js.CopyBytesToGo(evt, args[0])

		switch evt[0] {
		case HEARTBEAT:
			return js.ValueOf("HEARTBEAT")
		case HEARTBEAT_ACK:
			return js.ValueOf("HEARTBEAT_ACK")
		case ERROR:
			return js.ValueOf("ERROR")
		case DATA:
			return js.ValueOf("DATA")
		case DATA_ACK:
			return js.ValueOf("DATA_ACK")
		case REGISTER:
			return js.ValueOf("REGISTER")
		case AUTHENTICATE:
			return js.ValueOf("AUTHENTICATE")
		case SUB_GUILDS:
			return js.ValueOf("SUB_GUILDS")
		case ADD_USERS:
			return js.ValueOf("ADD_USERS")
		case REMOVE_USERS:
			return js.ValueOf("REMOVE_USERS")
		default:
			return js.ValueOf("UNKNOWN")
		}
	}

	marshalHeartbeatAck := func(this js.Value, args []js.Value) any {
		evt := []byte{HEARTBEAT_ACK}

		out := js.Global().Get("Uint8Array").New(len(evt))
		js.CopyBytesToJS(out, evt)
		return out
	}

	unmarshalError := func(this js.Value, args []js.Value) any {
		evt := make([]byte, args[0].Length())
		js.CopyBytesToGo(evt, args[0])

		return js.ValueOf(evt[1])
	}

	marshalData := func(this js.Value, args []js.Value) any {
		msg := make([]byte, args[1].Length())
		js.CopyBytesToGo(msg, args[1])

		evt := MarshalData(args[0].String(), msg)

		out := js.Global().Get("Uint8Array").New(len(evt))
		js.CopyBytesToJS(out, evt)
		return out
	}

	unmarshalData := func(this js.Value, args []js.Value) any {
		evt := make([]byte, args[0].Length())
		js.CopyBytesToGo(evt, args[0])

		guild, evtID, ts, msg := UnmarshalData(evt)

		msgOut := js.Global().Get("Uint8Array").New(len(msg))
		js.CopyBytesToJS(msgOut, msg)

		return js.ValueOf(map[string]interface{}{
			"guild": guild,
			"evt":   evtID,
			"ts":    ts,
			"msg":   msgOut,
		})
	}

	unmarshalDataAck := func(this js.Value, args []js.Value) any {
		evt := make([]byte, args[0].Length())
		js.CopyBytesToGo(evt, args[0])

		guild, evtID, ts := UnmarshalDataAck(evt)

		return js.ValueOf(map[string]interface{}{
			"guild": guild,
			"evt":   evtID,
			"ts":    ts,
		})
	}

	marshalRegister := func(this js.Value, args []js.Value) any {
		evt := append([]byte{REGISTER}, make([]byte, 16)...)

		out := js.Global().Get("Uint8Array").New(len(evt))
		js.CopyBytesToJS(out, evt)
		return out
	}

	unmarshalRegister := func(this js.Value, args []js.Value) any {
		evt := make([]byte, args[0].Length())
		js.CopyBytesToGo(evt, args[0])

		user, token := UnmarshalRegister(evt)

		tokenOut := js.Global().Get("Uint8Array").New(len(token))
		js.CopyBytesToJS(tokenOut, token)

		return js.ValueOf(map[string]interface{}{
			"user":  user,
			"token": token,
		})
	}

	marshalAuthenticate := func(this js.Value, args []js.Value) any {
		token := make([]byte, args[0].Length())
		js.CopyBytesToGo(token, args[0])

		out := append([]byte{AUTHENTICATE}, token...)

		evt := js.Global().Get("Uint8Array").New(len(out))
		js.CopyBytesToJS(evt, out)
		return evt
	}

	marshalSubGuilds := func(this js.Value, args []js.Value) any {
		guilds := make([]string, args[0].Length())
		for i := 0; i < args[0].Length(); i++ {
			guilds[i] = args[0].Index(i).String()
		}

		evt := MarshalSubGuilds(guilds)

		out := js.Global().Get("Uint8Array").New(len(evt))
		js.CopyBytesToJS(out, evt)
		return out
	}

	marshalAddUsers := func(this js.Value, args []js.Value) any {
		users := make([]string, args[1].Length())
		for i := 0; i < args[1].Length(); i++ {
			users[i] = args[1].Index(i).String()
		}

		evt := MarshalAddUsers(args[0].String(), users)

		out := js.Global().Get("Uint8Array").New(len(evt))
		js.CopyBytesToJS(out, evt)
		return out
	}

	marshalRemoveUsers := func(this js.Value, args []js.Value) any {
		users := make([]string, args[1].Length())
		for i := 0; i < args[1].Length(); i++ {
			users[i] = args[1].Index(i).String()
		}

		evt := MarshalAddUsers(args[0].String(), users)

		out := js.Global().Get("Uint8Array").New(len(evt))
		js.CopyBytesToJS(out, evt)
		return out
	}

	obj := js.ValueOf(map[string]interface{}{
		"eventType":           js.FuncOf(eventType),
		"marshalHeartbeatAck": js.FuncOf(marshalHeartbeatAck),
		"unmarshalError":      js.FuncOf(unmarshalError),
		"marshalData":         js.FuncOf(marshalData),
		"unmarshalData":       js.FuncOf(unmarshalData),
		"unmarshalDataAck":    js.FuncOf(unmarshalDataAck),
		"marshalRegister":     js.FuncOf(marshalRegister),
		"unmarshalRegister":   js.FuncOf(unmarshalRegister),
		"marshalAuthenticate": js.FuncOf(marshalAuthenticate),
		"marshalSubGuilds":    js.FuncOf(marshalSubGuilds),
		"marshalAddUsers":     js.FuncOf(marshalAddUsers),
		"marshalRemoveUsers":  js.FuncOf(marshalRemoveUsers),
	})

	return obj
}

// func populateRxMethods(rx *RxSession) js.Value {
// 	receive := func(this js.Value, args []js.Value) any {
// 		if len(args) != 1 {
// 			panic("sendMessage requires 1 parameter")
// 		}

// 		msg, err := base64.StdEncoding.DecodeString(args[0])
// 		if err != nil {
// 			panic(err)
// 		}

// 		b := new(bytes.Buffer)
// 		rx.ReceiveMessage(msg)

// 		return js.ValueOf(base64.StdEncoding.EncodeToString(b.Bytes()))
// 	}

// 	obj := js.ValueOf(map[string]interface{}{
// 		"receiveMessage": js.FuncOf(receive),
// 	})

// 	return obj
// }
