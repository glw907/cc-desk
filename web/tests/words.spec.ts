import { expect, test } from '@playwright/test';
import { dayWord, formatRange, formatTime } from '../src/lib/time';
import { tables, word, type Field } from '../src/lib/words';
import type { Pass, Round, Snapshot } from '../src/lib/snapshot';
import { allStates } from './state';

// Copied from the Go constants in internal/snapshot/snapshot.go and internal/state/types.go.
const written: Record<Field, string[]> = {
	passState: ['not_started', 'waiting', 'building', 'stopped', 'done'],
	roundState: ['not_designed', 'not_started', 'building', 'done'],
	stopReason: ['usage', 'irreversible', 'secret', 'auth_data', 'scope'],
	limit: ['two_pass_limit', 'usage'],
	needsKind: ['questions', 'nothing', 'not_designed'],
	guessOutcome: ['held', 'missed'],
	answerKind: ['ruling', 'carried'],
	nextSessionState: ['needs_you_now', 'ready_now', 'scheduled', 'nothing_waiting']
};

test('every written-out code has an entry', () => {
	const missing: string[] = [];
	for (const [field, codes] of Object.entries(written) as [Field, string[]][]) {
		for (const code of codes) if (!(code in tables[field])) missing.push(`${field}: ${code}`);
	}
	expect(missing, `no words for ${missing.join(', ')}`).toEqual([]);
});

test('every code in every printed snapshot has an entry', () => {
	const missing = new Set<string>();
	const check = (field: Field, code: string | null | undefined) => {
		if (code && !(code in tables[field])) missing.add(`${field}: ${code}`);
	};
	for (const s of Object.values(allStates()) as Snapshot[]) {
		check('nextSessionState', s.next_session.state);
		for (const rel of s.releases) {
			for (const r of rel.rounds as Round[]) {
				check('roundState', r.state);
				check('needsKind', r.needs.kind);
				for (const q of r.needs.questions) check('answerKind', q.answer_kind);
				for (const p of r.passes as Pass[]) {
					check('passState', p.state);
					check('stopReason', p.stop_reason);
					check('guessOutcome', p.guess_outcome);
					for (const q of p.questions) check('answerKind', q.answer_kind);
				}
			}
		}
		for (const w of s.waiting) for (const l of w.limits) check('limit', l);
	}
	expect([...missing], `no words for ${[...missing].join(', ')}`).toEqual([]);
});

test('the five pass words read as the design gives them', () => {
	expect(['not_started', 'waiting', 'building', 'stopped', 'done'].map((c) => word('passState', c))).toEqual([
		'Not started',
		'Waiting',
		'Building',
		'Stopped',
		'Done'
	]);
});

test('an unknown code shows with spaces for underscores', () => {
	expect(word('passState', 'half_way_there')).toBe('half way there');
});

test('times show in Alaska time to the minute', () => {
	expect(formatTime('2026-11-21T05:00:00Z')).toBe('8:00 pm');
	expect(formatTime('2026-11-21T09:00:00Z')).toBe('midnight');
	expect(formatTime('2026-11-20T21:00:00Z')).toBe('12:00 pm');
	expect(formatTime(null)).toBe('–');
	expect(formatRange('2026-11-21T03:00:00Z', '2026-11-21T06:30:00Z')).toBe('6:00 pm to 9:30 pm');
});

test('Today and Tomorrow compare Alaska dates', () => {
	const now = new Date('2026-11-20T23:00:00Z');
	expect(dayWord('2026-11-21T05:00:00Z', now)).toBe('Today');
	expect(dayWord('2026-11-21T09:30:00Z', now)).toBe('Tomorrow');
	expect(dayWord('2026-11-23T05:00:00Z', now)).toBeNull();
});
