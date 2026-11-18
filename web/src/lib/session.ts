// What the next decision session reads as, in one place so the widget, the front page's
// heading, and the browser tab's title always agree.
import type { NextSession, StateAnswer } from './snapshot';
import { formatWhen } from './time';
import { word } from './words';

/**
 * The state in words. A scheduled session reads as its time, never the word "Scheduled".
 * A state the page doesn't know shows as the code with spaces for underscores.
 */
export function sessionWords(next: NextSession, now: Date): string | null {
	if (next.state === null) return null;
	if (next.state === 'scheduled' && next.time !== null) return formatWhen(next.time, now, true);
	return word('nextSessionState', next.state);
}

/** The tab title: the widget's words then the app's name. */
export function tabTitle(answer: StateAnswer | null, now: Date): string {
	if (answer?.noState) return 'No state yet · cc-desk';
	const words = answer?.snapshot ? sessionWords(answer.snapshot.next_session, now) : null;
	return words === null ? 'cc-desk' : `${words} · cc-desk`;
}

/** The first letter in lower case, for words that follow a colon. */
export function lowerFirst(text: string): string {
	return text.charAt(0).toLowerCase() + text.slice(1);
}
