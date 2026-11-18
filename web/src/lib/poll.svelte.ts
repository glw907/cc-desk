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
	#tick = 0;
	#applied = 0;

	/** Fetches once now, then every `POLL_MS`. Returns the function that stops it. */
	start(): () => void {
		void this.refresh();
		this.#timer = setInterval(() => void this.refresh(), POLL_MS);
		return () => clearInterval(this.#timer);
	}

	/** Asks the server for the state once. A failed request leaves the last answer in place. */
	async refresh(): Promise<void> {
		const tick = ++this.#tick;
		const asked = new Date();
		this.now = asked;
		try {
			const res = await fetch('/api/state', { signal: AbortSignal.timeout(ANSWER_TIMEOUT_MS) });
			if (!res.ok) throw new Error(`state answered ${res.status}`);
			const answer = (await res.json()) as StateAnswer;
			// A slow answer to an older request must not overwrite a newer one.
			if (tick < this.#applied) return;
			this.#applied = tick;
			this.answer = answer;
			this.answeredAt = asked;
			this.connected = true;
		} catch {
			if (tick >= this.#applied) this.connected = false;
		}
	}
}

export const poll = new Poll();
