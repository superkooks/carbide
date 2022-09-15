<script setup lang="ts">
import { useGlobalStore } from "@/stores/global";
import { Go } from "@/assets/wasm_exec.js";
import { set as idbset, get as idbget } from "idb-keyval";

const store = useGlobalStore();

// const idbreq = window.indexedDB.open("spelt");

// idbreq.onupgradeneeded = () => {
//   idbreq.result.createObjectStore("global");
// };

// idbreq.onerror = () => {
//   console.log("failed to load database");
//   return;
// };

// idbreq.onsuccess = () => {
//   const idb = idbreq.result;
//   const tx = idb.transaction("global", "readwrite");

//   // Hydrate store
//   const objstore = tx.objectStore("global");

//   const req = objstore.get("state");
//   req.onsuccess = () => {
//     const newState = req.result;
//     if (newState != undefined) {
//       store.overwrite(JSON.parse(newState));
//     }
//   };

//   // Listen for updates
//   store.$subscribe((mutation, state) => {
//     console.log(JSON.stringify(state));
//     const objstore = idb
//       .transaction("global", "readwrite")
//       .objectStore("global");
//     objstore.put(JSON.stringify(state), "state");
//   });
// };

idbget("state").then((v) => {
  store.overwrite(JSON.parse(v as string));
});

// Listen for updates
store.$subscribe((mutation, state) => {
  console.log(JSON.stringify(state));
  idbset("state", JSON.stringify(state));
});

const go = new Go();
WebAssembly.instantiateStreaming(
  fetch("/buckwheat.wasm"),
  go.importObject
).then((result) => {
  go.run(result.instance);
  const txs = window.buckwheat.doubleTx();
  const alice = txs[0];
  const bob = txs[1];

  const a2 = window.buckwheat.importTx(alice.export());

  const aup = a2.generateUpdate();
  bob.receiveMessage(aup);

  const bup = bob.generateUpdate();
  a2.receiveMessage(bup);

  const ctext = a2.sendMessage(btoa("testing 123"));
  console.log(ctext);
  console.log(bob.receiveMessage(ctext));
});
</script>
