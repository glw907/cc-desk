// Every time on the page shows in Alaska time, to the minute, whatever zone the browser is in.
const ZONE = 'America/Anchorage';

/** The screen-reader text for the dash that stands for a missing time. */
export const NONE = 'None';

/** The dash shown for a missing time. Pair it with `NONE` as hidden text. */
export const DASH = '–';

const clock = new Intl.DateTimeFormat('en-US', {
	timeZone: ZONE,
	hour: 'numeric',
	minute: '2-digit',
	hourCycle: 'h12'
});

const day = new Intl.DateTimeFormat('en-CA', {
	timeZone: ZONE,
	year: 'numeric',
	month: '2-digit',
	day: '2-digit'
});

/** "8:00 pm", with 12:00 am as "midnight". */
export function formatTime(iso: string | null): string {
	if (iso === null) return DASH;
	const parts = Object.fromEntries(
		clock.formatToParts(new Date(iso)).map((p) => [p.type, p.value])
	);
	if (parts.hour === '12' && parts.minute === '00' && parts.dayPeriod === 'AM') return 'midnight';
	return `${parts.hour}:${parts.minute} ${parts.dayPeriod.toLowerCase()}`;
}

/** Two bare times joined by "to". */
export function formatRange(low: string, high: string): string {
	return `${formatTime(low)} to ${formatTime(high)}`;
}

function alaskaDate(d: Date): string {
	return day.format(d);
}

/** "Today" or "Tomorrow" when the time's Alaska date is the clock's or the next one; else null. */
export function dayWord(iso: string, now: Date): 'Today' | 'Tomorrow' | null {
	const date = alaskaDate(new Date(iso));
	if (date === alaskaDate(now)) return 'Today';
	const next = new Date(now.getTime() + 24 * 60 * 60 * 1000);
	if (date === alaskaDate(next)) return 'Tomorrow';
	return null;
}
