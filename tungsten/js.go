package main

import (
	"bytes"
	"syscall/js"

	"github.com/google/uuid"
)

// The signatures of all of these functions are available under
// tungsten in /app/src/assets/wasm_exec.ts

// This file is my least favourite.

func genGlobalJS() {
	obj := js.ValueOf(map[string]interface{}{})

	obj.Set("genTx", js.FuncOf(genTxWrapped))
	obj.Set("importTx", js.FuncOf(importTxWrapped))

	// TODO: Remove temp functions
	obj.Set("doubleTx", js.FuncOf(doubleTxWrapped))

	js.Global().Set("tungsten", obj)
}

func genTxWrapped(this js.Value, args []js.Value) any {
	id, err := uuid.Parse(args[0].String())
	if err != nil {
		panic(err)
	}

	return populateTxMethods(GenTx(id))
}

func importTxWrapped(this js.Value, args []js.Value) any {
	buf := make([]byte, args[0].Length())
	js.CopyBytesToGo(buf, args[0])

	return populateTxMethods(ImportTx(bytes.NewBuffer(buf)))
}

// TODO: Remove temp function
func doubleTxWrapped(this js.Value, args []js.Value) any {
	alice := GenTx(uuid.New())
	bob := GenTx(uuid.New())

	RxFromTx(bob, alice)
	RxFromTx(alice, bob)

	arr := js.ValueOf([]interface{}{})
	arr.SetIndex(0, populateTxMethods(alice))
	arr.SetIndex(1, populateTxMethods(bob))

	return arr
}

func populateTxMethods(tx *TxSession) js.Value {
	send := func(this js.Value, args []js.Value) any {
		msg := make([]byte, args[1].Length())
		js.CopyBytesToGo(msg, args[1])

		ratchetID, err := uuid.Parse(args[0].String())
		if err != nil {
			panic(err)
		}

		b := new(bytes.Buffer)
		tx.SendMessage(ratchetID, msg, b)

		out := js.Global().Get("Uint8Array").New(b.Len())
		js.CopyBytesToJS(out, b.Bytes())
		return out
	}

	receive := func(this js.Value, args []js.Value) any {
		in := make([]byte, args[0].Length())
		js.CopyBytesToGo(in, args[0])

		msg, err := tx.ReceiveMessage(in)
		outBytes := js.Global().Get("Uint8Array").New(len(msg))
		js.CopyBytesToJS(outBytes, msg)

		return js.ValueOf(map[string]interface{}{
			"msg":   outBytes,
			"error": err != nil,
		})
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
