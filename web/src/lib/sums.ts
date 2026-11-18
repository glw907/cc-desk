// The sums the sidebar and the boards display, counted from the snapshot's rounds and passes.
// Estimates, times, and rule outcomes are never worked out here: they come from the snapshot.
import type { Pass, Release, Round, Snapshot } from './snapshot';

/** A part of a scale bar: a share of the work and the state that colors it. */
export interface Part {
	minutes: number;
	state: string;
	asks: boolean;
}

/** How many of the parts are `done`. */
export function doneCount(parts: { state: string }[]): number {
	return parts.filter((p) => p.state === 'done').length;
}

/** The percent of planned minutes in `done` parts, rounded to a whole number. */
export function percentDone(parts: { state: string; planned_minutes: number }[]): number {
	const all = parts.reduce((sum, p) => sum + p.planned_minutes, 0);
	if (all === 0) return 0;
	const done = parts.filter((p) => p.state === 'done').reduce((sum, p) => sum + p.planned_minutes, 0);
	return Math.round((100 * done) / all);
}

/**
 * Whether a round needs Geoff: it is not done, and it is either not designed yet or has a
 * question nobody has answered.
 */
export function needsGeoff(round: Round): boolean {
	if (round.state === 'done') return false;
	if (round.needs.kind === 'not_designed') return true;
	return round.needs.kind === 'questions' && round.needs.questions.some((q) => !q.answered);
}

/** The round a `project/round` name points at, or undefined when the snapshot has no such round. */
export function findRound(snapshot: Snapshot, name: string): Round | undefined {
	const [project, round] = name.split('/');
	return snapshot.releases.find((r) => r.project === project)?.rounds.find((r) => r.name === round);
}

/** A round with the release that holds it. */
export interface RoundFound {
	release: Release;
	round: Round;
}

/** A pass with the round and release that hold it. */
export interface PassFound extends RoundFound {
	pass: Pass;
}

/**
 * The round at `/<project>/<round>`. Two projects may each have a round of one name, so the
 * project picks the release first.
 */
export function findRoundAt(snapshot: Snapshot, project: string, name: string): RoundFound | undefined {
	for (const release of snapshot.releases) {
		if (release.project !== project) continue;
		const round = release.rounds.find((r) => r.name === name);
		if (round) return { release, round };
	}
	return undefined;
}

/**
 * The pass at `/<project>/<round>/<pass>`. A pass name is unique within its release, but the
 * address must also name the pass's own round, or the pass is not found.
 */
export function findPassAt(
	snapshot: Snapshot,
	project: string,
	roundName: string,
	name: string
): PassFound | undefined {
	const found = findRoundAt(snapshot, project, roundName);
	const pass = found?.round.passes.find((p) => p.name === name);
	return found && pass ? { ...found, pass } : undefined;
}

/** The text of a question by its id, looked up in the project's rounds and passes. */
export function questionText(snapshot: Snapshot, project: string, id: string): string | undefined {
	for (const release of snapshot.releases) {
		if (release.project !== project) continue;
		for (const round of release.rounds) {
			const all = [...round.needs.questions, ...round.passes.flatMap((p) => p.questions)];
			const hit = all.find((q) => q.id === id);
			if (hit) return hit.text;
		}
	}
	return undefined;
}

/** The rounds of a release with a pass building. */
export function buildingRounds(release: Release): Round[] {
	return release.rounds.filter((r) => r.state === 'building');
}

/** A round is opened to its passes on a board when it is building or a pass waits or is stopped. */
export function isOpen(round: Round): boolean {
	return (
		round.state === 'building' ||
		round.passes.some((p) => p.state === 'waiting' || p.state === 'stopped')
	);
}

/** The scale bar parts for a release's rounds. */
export function roundParts(rounds: Round[]): Part[] {
	return rounds.map((r) => ({ minutes: r.planned_minutes, state: r.state, asks: needsGeoff(r) }));
}

/** The scale bar parts for a round's passes. */
export function passParts(passes: Pass[]): Part[] {
	return passes.map((p) => ({ minutes: p.planned_minutes, state: p.state, asks: false }));
}

/** "1 question", "2 questions". */
export function count(n: number, one: string, many = `${one}s`): string {
	return `${n} ${n === 1 ? one : many}`;
}

/** "alpha", "alpha and beta", "alpha, beta, and gamma". */
export function list(items: string[]): string {
	if (items.length < 3) return items.join(' and ');
	return `${items.slice(0, -1).join(', ')}, and ${items[items.length - 1]}`;
}
