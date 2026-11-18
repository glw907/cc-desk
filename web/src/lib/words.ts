// One table words every code the snapshot carries. The five pass words and the widget states are
// the design's. An unknown code still shows, with its underscores turned to spaces, so a code the
// Sessions round adds reads as plain text instead of breaking the page.

const passStates: Record<string, string> = {
	not_started: 'Not started',
	waiting: 'Waiting',
	building: 'Building',
	stopped: 'Stopped',
	done: 'Done'
};

const roundStates: Record<string, string> = {
	not_designed: 'Not designed yet',
	not_started: 'Not started',
	building: 'Building',
	done: 'Done'
};

const stopReasons: Record<string, string> = {
	usage: 'Usage limit reached',
	irreversible: 'An irreversible step',
	secret: 'A secret is needed',
	auth_data: 'A sign-in or data change',
	scope: 'A change of scope',
	permission_prompt: 'A permission prompt',
	sign_in_expired: 'The sign-in ran out',
	unreadable_stop: 'The stop could not be read',
	relaunch_limit: 'Restarted twice already',
	launch_failed: 'The session would not start'
};

const sessionStates: Record<string, string> = {
	working: 'Working',
	blocked: 'Waiting on a prompt',
	idle: 'Idle',
	missing: 'Not in the session list'
};

const checkDecisions: Record<string, string> = {
	none: 'did nothing',
	end: 'ended the session',
	relaunch: 'restarted the session',
	end_and_relaunch: 'ended the session and restarted it',
	launch_again: 'tried the launch again',
	launch: 'started a session',
	stop: 'stopped the pass',
	wait_usage: 'waited for the usage limit',
	wait_link: 'waited for a terminal link'
};

const checkReasons: Record<string, string> = {
	lost_connection: 'the connection was lost',
	idle: 'the session was idle',
	missing: 'the session was not in the list',
	...Object.fromEntries(Object.entries(stopReasons).map(([code, text]) => [code, text.toLowerCase()]))
};

const limits: Record<string, string> = {
	two_pass_limit: 'Two passes already running',
	usage: 'Usage limit reached'
};

const needsKinds: Record<string, string> = {
	questions: 'Questions to answer',
	nothing: 'Nothing needed',
	not_designed: 'Not designed yet'
};

const guessOutcomes: Record<string, string> = {
	held: 'Guess held',
	missed: 'Guess missed'
};

const answerKinds: Record<string, string> = {
	ruling: 'Ruled on',
	carried: 'Carried over'
};

const nextSessionStates: Record<string, string> = {
	needs_you_now: 'Needs you now',
	ready_now: 'Ready now',
	scheduled: 'Scheduled',
	nothing_waiting: 'Nothing waiting'
};

/** Every table, keyed by the field its codes come from. The words test reads this. */
export const tables = {
	passState: passStates,
	roundState: roundStates,
	stopReason: stopReasons,
	sessionState: sessionStates,
	checkDecision: checkDecisions,
	checkReason: checkReasons,
	limit: limits,
	needsKind: needsKinds,
	guessOutcome: guessOutcomes,
	answerKind: answerKinds,
	nextSessionState: nextSessionStates
};

export type Field = keyof typeof tables;

/** The words for a code. An unknown code shows as itself with spaces for underscores. */
export function word(field: Field, code: string): string {
	return tables[field][code] ?? code.replaceAll('_', ' ');
}
