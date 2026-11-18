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
	scope: 'A change of scope'
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
