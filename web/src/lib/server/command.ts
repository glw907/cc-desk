// The one place the server runs a command. "Mark as read" calls the cc-desk binary, which is
// the only writer of the state folder: this file opens no file and starts no shell.
import { execFile } from 'node:child_process';
import path from 'node:path';
import { stateDir } from './state.ts';

/** The ids `cc-desk` gives a report. The route's matcher checks the same pattern. */
const REPORT_ID = /^rep-\d{4,}$/;

/** Longer than the binary's 10 second wait for the state lock. */
const TIMEOUT_MS = 30_000;

/** Starts every message the page shows when the click could not be recorded. */
export const NOT_RECORDED = "Mark as read couldn't be recorded:";

/** `ok`, or the sentence to show. */
export type MarkResult = { ok: true } | { ok: false; message: string };

const firstLine = (text: string) => text.trim().split('\n')[0] ?? '';

/**
 * Records a report as read with `cc-desk report read`. The binary comes from `CC_DESK_BINARY`
 * and must be an absolute path: an unset or relative value is a failure, never a `PATH` lookup.
 * Each argument is its own array member, and an id that is not `rep-` and digits is refused
 * before any command starts.
 */
export function markRead(id: string, env: NodeJS.ProcessEnv = process.env): Promise<MarkResult> {
	const binary = env.CC_DESK_BINARY;
	if (!REPORT_ID.test(id)) return Promise.resolve({ ok: false, message: `${NOT_RECORDED} that is not a report.` });
	if (!binary || !path.isAbsolute(binary)) {
		return Promise.resolve({ ok: false, message: `${NOT_RECORDED} the page does not know where cc-desk is.` });
	}
	const args = ['--state-dir', stateDir(env), 'report', 'read', id];
	return new Promise((resolve) => {
		execFile(binary, args, { timeout: TIMEOUT_MS, env }, (err, _stdout, stderr) => {
			if (!err) return resolve({ ok: true });
			const why = firstLine(stderr) || (err.killed ? 'the command took too long.' : firstLine(err.message));
			resolve({ ok: false, message: `${NOT_RECORDED} ${why}` });
		});
	});
}
