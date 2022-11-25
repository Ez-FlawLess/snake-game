const goWasm = new Go()

function keyDownListener(event) {
    listenToKeys(event.key)
}

WebAssembly.instantiateStreaming(fetch("main.wasm"), goWasm.importObject)
    .then((result) => {
        goWasm.run(result.instance)

        document.addEventListener('keydown', keyDownListener, false);
        
    })