package main

import (
	"bytes"
	"encoding/base64"
	"syscall/js"
)

// The signatures of all these functions are available under
// buckwheat in /app/src/assets/wasm_exec.ts

func genGlobalJS() {
	obj := js.ValueOf(map[string]interface{}{})

	obj.Set("genTx", js.FuncOf(genTxWrapped))
	obj.Set("importTx", js.FuncOf(importTxWrapped))

	// TODO: Remove temp functions
	obj.Set("doubleTx", js.FuncOf(doubleTxWrapped))

	js.Global().Set("buckwheat", obj)
}

func genTxWrapped(this js.Value, args []js.Value) any {
	if len(args) > 0 {
		panic("genTX requires 0 parameters")
	}

	return populateTxMethods(GenTx())
}

func importTxWrapped(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		panic("importTX requires 1 parameter")
	}

	msg, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		panic(err)
	}

	return populateTxMethods(ImportTx(bytes.NewBuffer(msg)))
}

// TODO: Remove temp function
func doubleTxWrapped(this js.Value, args []js.Value) any {
	if len(args) > 0 {
		panic("rxFromTX requires 0 parameter")
	}

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
		if len(args) != 1 {
			panic("sendMessage requires 1 parameter")
		}

		msg, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			panic(err)
		}

		b := new(bytes.Buffer)
		tx.SendMessage(msg, b)

		return js.ValueOf(base64.StdEncoding.EncodeToString(b.Bytes()))
	}

	receive := func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			panic("sendMessage requires 1 parameter")
		}

		msg, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			panic(err)
		}

		return js.ValueOf(base64.StdEncoding.EncodeToString(tx.ReceiveMessage(msg)))
	}

	genUpdate := func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			panic("generateUpdate requires 0 parameters")
		}

		b := new(bytes.Buffer)
		tx.GenerateUpdate(b)

		return js.ValueOf(base64.StdEncoding.EncodeToString(b.Bytes()))
	}

	export := func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			panic("export requires 0 parameters")
		}

		b := new(bytes.Buffer)
		tx.Export(b)

		return js.ValueOf(base64.StdEncoding.EncodeToString(b.Bytes()))
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
