const { execSync } = require('child_process');

console.log('Building Go WebAssembly binary...');

// Set environment variables for compilation
const env = {
  ...process.env,
  GOOS: 'wasip1',
  GOARCH: 'wasm'
};

// 1. Run go build
try {
  execSync('go build -trimpath -ldflags="-s -w" -o main.wasm .', { env, stdio: 'inherit' });
  console.log('Go build successful.');
} catch (err) {
  console.error('Go build failed:', err.message);
  process.exit(1);
}

// 2. Run wasm-opt
try {
  console.log('Optimizing Wasm size using wasm-opt...');
  execSync('npx wasm-opt -Oz --enable-bulk-memory --enable-nontrapping-float-to-int -o main.wasm main.wasm', { stdio: 'inherit' });
  console.log('Optimization successful.');
} catch (err) {
  console.error('wasm-opt optimization failed (skipping):', err.message);
  // We do not fail the build if wasm-opt fails
}
