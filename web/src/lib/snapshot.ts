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
	/** The id of the ruling this one replaces as a question's answer. */
	replaces?: string;
}

export interface Question {
	id: string;
	text: string;
	options: string[];
	recommended: string;
	answered: boolean;
	answer_kind: string | null;
	background: string[];
	causes: Record<string, string>;
	reason: string | null;
	never_guess: boolean;
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
	/** When the open stop began; null unless the pass is stopped. */
	stopped_at: Time | null;
	/** `questions` when the pass rests on any, else `nothing`. */
	needs_kind: string;
	/** The pass's own reason it needs nothing, else its round's, else null. */
	needs_reason: string | null;
	/** Set for a pass a ruling closed without a merge. */
	closed: Closed | null;
}

/** The ruling that closed a pass, and when. */
export interface Closed {
	ruling: string;
	time: Time;
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

/** A permission prompt a session stopped on, as `internal/state/types.go` records it. */
export interface Prompt {
	time: Time;
	tool: string;
	command?: string;
	suggested_rule?: string;
	answer?: string;
}

/** A call a session's settings turned away, as `internal/state/types.go` records it. */
export interface Refusal {
	time: Time;
	tool: string;
	target: string;
	reason: string;
}

/** One session record of a live release. Pass is null for a conductor and a design writer; round is null for a conductor. */
export interface SessionEntry {
	project: string;
	round: string | null;
	pass: string | null;
	kind: string;
	name: string;
	short_id: string;
	last_state: string | null;
	last_seen: Time | null;
	last_working: Time | null;
	relaunches: number;
	stop_reason: string | null;
	prompts: Prompt[];
	/** When the session's stop began; null when it is not stopped. */
	stopped_at: Time | null;
	refusals: Refusal[];
	/** The note a merged pass's cleanup left: what stayed in place. */
	cleanup?: string;
}

/** One cause that needs Geoff now. */
export interface NeedEntry {
	project: string;
	round: string | null;
	session: string | null;
	reason: string;
	cause: string;
	todo: string;
}

/** One decision the session check made on a record. */
export interface Decision {
	project: string;
	record: string;
	session: string;
	decision: string;
	reason?: string;
	detail?: string;
	counted?: boolean;
	outcome?: string;
}

/** The last run of the session check. */
export interface LastCheck {
	time: Time;
	list_failures: number;
	decisions: Decision[];
}

/** One thing the session check did: a stop it applied or an outcome it carried out. */
export interface CheckLogLine {
	format: number;
	time: Time;
	project: string;
	record: string;
	decision: string;
	outcome: string;
	reason?: string;
	detail?: string;
}

/** One decision a pass or round made, and the other way it could have gone. */
export interface Call {
	call: string;
	chosen: string;
	other: string;
	why: string;
}

/** Something the plan missed, and what it cost. */
export interface Miss {
	miss: string;
	cost: string;
}

/** One pass in a report's folded detail. */
export interface ReportPass {
	name: string;
	state: string;
	planned_minutes: number;
	actual_minutes: number | null;
	pull_request: number | null;
	closed: Closed | null;
}

/** A report from a round or a release. `read` is null until it is marked read. */
export interface Report {
	id: string;
	project: string;
	/** Null for a release's report. */
	round: string | null;
	title: string;
	summary: string;
	calls: Call[];
	misses: Miss[];
	sent: Time;
	read: Time | null;
	passes: ReportPass[];
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
	sessions: SessionEntry[];
	needs_you_now: NeedEntry[];
	last_check: LastCheck | null;
	/** The last 50 lines of the check log, oldest first. */
	check_log: CheckLogLine[];
	/** Every report, oldest first. */
	reports: Report[];
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
