// Builds the real cc-desk binary once per run. Tests use it to print the state they draw,
// so a state in a test is always what the binary writes, never a hand-built snapshot.
import { execFileSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';

export default function globalSetup(): void {
	const repoRoot = path.resolve(import.meta.dirname, '..', '..');
	const dir = mkdtempSync(path.join(tmpdir(), 'cc-desk-bin-'));
	const bin = path.join(dir, 'cc-desk');
	execFileSync('go', ['build', '-o', bin, './cmd/cc-desk'], { cwd: repoRoot, stdio: 'inherit' });
	process.env.CC_DESK_BINARY = bin;
}
