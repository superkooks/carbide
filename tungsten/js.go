package main

import (
	"bytes"
	"syscall/js"
)

// The signatures of all these functions are available under
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
	marshalData := func(this js.Value, args []js.Value) any {
		msg := make([]byte, args[1].Length())
		js.CopyBytesToGo(msg, args[1])

		buf := MarshalData(args[0].String(), msg)

		out := js.Global().Get("Uint8Array").New(len(buf))
		js.CopyBytesToJS(out, buf)
		return out
	}

	marshalSubGuilds := func(this js.Value, args []js.Value) any {
		guilds := make([]string, args[0].Length())
		for i := 0; i < args[0].Length(); i++ {
			guilds[i] = args[0].Index(i).String()
		}

		buf := MarshalSubGuilds(guilds)

		out := js.Global().Get("Uint8Array").New(len(buf))
		js.CopyBytesToJS(out, buf)
		return out
	}

	obj := js.ValueOf(map[string]interface{}{
		"marshalData":      js.FuncOf(marshalData),
		"marshalSubGuilds": js.FuncOf(marshalSubGuilds),
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
