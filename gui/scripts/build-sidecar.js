// Build the CLI that the GUI application embeds.
//
// The Rust crate names the file with a Rust target triple (see `TAI_BIN` in lib.rs), and the
// binary is gitignored, so this has to run before the crate will compile at all - which is
// why `pretauri` chains it in.
//
// Set GOOS/GOARCH to cross-compile; the triple and file extension follow from them, so the
// name always matches what the crate is looking for.

import { execFileSync } from 'node:child_process';
import { mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');

const goos = process.env.GOOS || (process.platform === 'win32' ? 'windows' : process.platform);
const goarch = process.env.GOARCH || (process.arch === 'arm64' ? 'arm64' : 'amd64');

const architectures = { amd64: 'x86_64', arm64: 'aarch64' };
const platforms = {
	windows: 'pc-windows-msvc',
	linux: 'unknown-linux-gnu',
	darwin: 'apple-darwin'
};

const architecture = architectures[goarch];
const platform = platforms[goos];
if (!architecture || !platform) {
	console.error(`Unsupported target: GOOS=${goos} GOARCH=${goarch}`);
	process.exit(1);
}

const name = `tai-${architecture}-${platform}${goos === 'windows' ? '.exe' : ''}`;
const output = join(root, 'src-tauri', 'binaries', name);

mkdirSync(dirname(output), { recursive: true });
console.log(`Building ${name}`);
execFileSync('go', ['build', '-o', output, '../tai'], {
	cwd: root,
	env: { ...process.env, GOOS: goos, GOARCH: goarch },
	stdio: 'inherit'
});
