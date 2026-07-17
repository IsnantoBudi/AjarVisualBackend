const { execSync } = require('child_process');

console.log('Building Go WebAssembly binary...');

// Set environment variables for compilation
const env = {
  ...process.env,
  GOOS: 'js',
  GOARCH: 'wasm'
};

// 1. Try to compile with TinyGo, fallback to Go if it fails
let buildSuccess = false;

try {
  console.log('Attempting to build with TinyGo...');
  execSync('tinygo build -o main.wasm -target=wasm -no-debug .', { env, stdio: 'inherit' });
  console.log('TinyGo build successful.');
  buildSuccess = true;
} catch (err) {
  console.warn('TinyGo build failed or not available, falling back to standard Go build...', err.message);
}

if (!buildSuccess) {
  try {
    console.log('Building with standard Go...');
    execSync('go build -trimpath -ldflags="-s -w" -o main.wasm .', { env, stdio: 'inherit' });
    console.log('Go build successful.');
  } catch (err) {
    console.error('Go build failed:', err.message);
    process.exit(1);
  }
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
