import './wasm_exec.js';
import wasm from './main.wasm';

const go = new Go();
const instance = new WebAssembly.Instance(wasm, go.importObject);
go.run(instance);

export default {
  async fetch(request, env, ctx) {
    return globalThis.workers.fetch(request, env, ctx);
  }
};
