// Reads the state folder. The server never opens a file for writing: the cc-desk binary is the
// only writer, and this module only reads what it leaves. The one command the server runs,
// "Mark as read", lives in command.ts.
import { readFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import type { Snapshot, StateAnswer } from '../snapshot';

/** Starts the text shown for a snapshot.json that isn't JSON; the parser's message follows. */
export const NOT_JSON = "snapshot.json isn't valid JSON:";

let lastGood: Snapshot | null = null;

/** The state folder: the test override, else `$XDG_STATE_HOME/cc-desk`, else `~/.local/state/cc-desk`. */
export function stateDir(env: NodeJS.ProcessEnv = process.env): string {
	if (env.CC_DESK_STATE_DIR) return env.CC_DESK_STATE_DIR;
	if (env.XDG_STATE_HOME) return path.join(env.XDG_STATE_HOME, 'cc-desk');
	return path.join(os.homedir(), '.local', 'state', 'cc-desk');
}

async function readOrNull(file: string): Promise<string | null> {
	try {
		return await readFile(file, 'utf8');
	} catch (err) {
		if ((err as NodeJS.ErrnoException).code === 'ENOENT') return null;
		throw err;
	}
}

/**
 * Reads both files afresh. A snapshot.json that is missing is "No state yet", whatever was read
 * before; one that isn't JSON keeps the last good snapshot and adds the parse error beside it.
 */
export async function readState(dir: string = stateDir()): Promise<StateAnswer> {
	const [snapshotText, errorText] = await Promise.all([
		readOrNull(path.join(dir, 'snapshot.json')),
		readOrNull(path.join(dir, 'snapshot-error.json'))
	]);

	let error: string | null = null;
	if (errorText !== null) {
		try {
			const parsed = JSON.parse(errorText) as { message?: string };
			error = parsed.message ?? errorText;
		} catch {
			error = errorText;
		}
	}

	if (snapshotText === null) {
		lastGood = null;
	} else {
		try {
			lastGood = JSON.parse(snapshotText) as Snapshot;
		} catch (err) {
			error = `${NOT_JSON} ${(err as Error).message}`;
		}
	}

	return {
		snapshot: lastGood,
		error,
		lastGood: lastGood?.generated_at ?? null,
		noState: lastGood === null
	};
}
