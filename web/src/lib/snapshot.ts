// The shape of snapshot.json, mirroring internal/snapshot/snapshot.go and the ruling in
// internal/state/types.go. A Go pointer is `| null`; an omitempty field is optional. Codes
// are plain strings so a code the Sessions round adds still type-checks and reaches words.ts.

/** An ISO 8601 time in UTC, as the binary writes it. */
export type Time = string;

export interface Ruling {
	format: number;
	id: string;
	time: Time;
	desk: string;
	words: string;
	project?: string;
	round?: string;
	question?: string;
	option?: string;
}

export interface Question {
	id: string;
	text: string;
	options: string[];
	recommended: string;
	answered: boolean;
	answer_kind: string | null;
}

export interface Needs {
	kind: string;
	reason: string | null;
	questions: Question[];
}

export interface Pass {
	name: string;
	summary: string;
	state: string;
	planned_minutes: number;
	actual_minutes: number | null;
	started: Time | null;
	finished: Time | null;
	stop_reason: string | null;
	questions: Question[];
	minutes_before: number | null;
	needed_time: Time | null;
	free: boolean | null;
	why_not: string[];
	guess_outcome: string | null;
}

export interface Round {
	name: string;
	summary: string;
	state: string;
	after: string[];
	planned_minutes: number;
	actual_minutes: number | null;
	started: Time | null;
	finished: Time | null;
	needs: Needs;
	minutes_before: number | null;
	needed_time: Time | null;
	forecast_time: Time | null;
	passes: Pass[];
}

export interface Release {
	project: string;
	name: string;
	repo: string;
	pace: number | null;
	pace_low: number | null;
	pace_high: number | null;
	rough: boolean | null;
	rounds: Round[];
}

export interface Range {
	low: Time;
	high: Time;
}

export interface NextSession {
	state: string | null;
	time: Time | null;
	range: Range | null;
	rough: boolean | null;
	question_count: number | null;
	rounds: string[];
	projects: string[];
	not_designed: string[];
	waiting_on_design: string[];
	reasons: string[];
}

export interface WaitEntry {
	project: string;
	round: string;
	questions: number;
}

export interface WaitingPass {
	project: string;
	pass: string;
	limits: string[];
	since: Time;
	reset_text: string | null;
}

export interface OpenGuess {
	project: string;
	pass: string;
	question: string;
	minutes: number;
}

export interface Snapshot {
	format: number;
	generated_at: Time;
	releases: Release[];
	rulings: Ruling[];
	next_session: NextSession;
	waits_for_geoff: WaitEntry[];
	waiting: WaitingPass[];
	open_guesses: OpenGuess[];
	guess_minutes_total: number;
	guess_minutes_limit: number;
}

/** What `/api/state` answers: everything the page draws and every banner it may show. */
export interface StateAnswer {
	/** The snapshot to draw: the file's, or the last good one. Null before any good read. */
	snapshot: Snapshot | null;
	/** The error file's message, or the app's own text for a snapshot.json that isn't JSON. */
	error: string | null;
	/** The `generated_at` of the last good snapshot the server read. */
	lastGood: Time | null;
	/** True when no snapshot.json has ever been read. */
	noState: boolean;
}
