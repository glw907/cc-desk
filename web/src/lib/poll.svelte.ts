import type { StateAnswer } from './snapshot';

/** How often the page asks the server for the state. */
export const POLL_MS = 2000;

/** A slow answer counts as no answer after this long, so one hung request never stalls the poll. */
const ANSWER_TIMEOUT_MS = 5000;

/**
 * The page's copy of the state. The page is browser-only, so one shared instance is safe.
 */
class Poll {
	answer = $state<StateAnswer | null>(null);

	/**
	 * Whether the last request got an answer. It comes from the request's outcome, never from
	 * comparing dates: a pinned test clock freezes `Date.now()` while the timers still run.
	 */
	connected = $state(true);

	/** The browser's clock when the last good answer arrived. */
	answeredAt = $state<Date | null>(null);

	/** The browser's clock at the last tick. "Today" and "20 minutes ago" are measured from it. */
	now = $state(new Date());

	#timer: ReturnType<typeof setInterval> | undefined;

	/** Fetches once now, then every `POLL_MS`. Returns the function that stops it. */
	start(): () => void {
		void this.refresh();
		this.#timer = setInterval(() => void this.refresh(), POLL_MS);
		return () => clearInterval(this.#timer);
	}

	/** Asks the server for the state once. A failed request leaves the last answer in place. */
	async refresh(): Promise<void> {
		this.now = new Date();
		try {
			const res = await fetch('/api/state', { signal: AbortSignal.timeout(ANSWER_TIMEOUT_MS) });
			if (!res.ok) throw new Error(`state answered ${res.status}`);
			this.answer = (await res.json()) as StateAnswer;
			this.answeredAt = this.now;
			this.connected = true;
		} catch {
			this.connected = false;
		}
	}
}

export const poll = new Poll();
