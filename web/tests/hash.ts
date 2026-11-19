// A fingerprint of a folder, so a test can show which files a request changed.
import { createHash } from 'node:crypto';
import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';

/** SHA-256 of every file under `root`, by path relative to `root`. */
export function hashTree(root: string): Record<string, string> {
	const out: Record<string, string> = {};
	const walk = (dir: string) => {
		for (const entry of readdirSync(dir, { withFileTypes: true })) {
			const full = path.join(dir, entry.name);
			if (entry.isDirectory()) walk(full);
			else out[path.relative(root, full)] = createHash('sha256').update(readFileSync(full)).digest('hex');
		}
	};
	walk(root);
	return out;
}
