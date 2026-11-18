// Prints every state the tests draw, then type-checks each against src/lib/snapshot.ts. A field
// the binary prints that the type lacks fails as TS2353; one the type needs and the binary omits
// fails as TS2741. The check is a typed const, not `satisfies`: TS 6 words a missing property
// under `satisfies` as TS1360.
import { execFileSync } from 'node:child_process';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { expect, test } from '@playwright/test';
import { allStates } from './state';

const webRoot = path.resolve(import.meta.dirname, '..');
const typeModule = path.join(webRoot, 'src', 'lib', 'snapshot');

test('every snapshot the binary prints satisfies the Snapshot type', () => {
	const dir = mkdtempSync(path.join(tmpdir(), 'cc-desk-types-'));
	const files: string[] = [];
	for (const [name, snapshot] of Object.entries(allStates())) {
		const file = path.join(dir, `${name}.ts`);
		writeFileSync(
			file,
			`import type { Snapshot } from ${JSON.stringify(typeModule)};\nexport const snapshot: Snapshot = ${JSON.stringify(snapshot, null, 1)};\n`
		);
		files.push(file);
	}
	writeFileSync(
		path.join(dir, 'tsconfig.json'),
		JSON.stringify({
			compilerOptions: {
				strict: true,
				noEmit: true,
				module: 'esnext',
				moduleResolution: 'bundler',
				target: 'es2022',
				skipLibCheck: true
			},
			files
		})
	);
	let output = '';
	try {
		execFileSync(path.join(webRoot, 'node_modules', '.bin', 'tsc'), ['-p', dir], {
			encoding: 'utf8'
		});
	} catch (err) {
		output = (err as { stdout?: string }).stdout ?? String(err);
	}
	expect(output, `tsc said:\n${output}`).toBe('');
});
