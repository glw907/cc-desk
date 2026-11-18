// Color never carries a state alone, and the app's CSS keeps to Survey's tokens. Every word, name,
// and pattern below is written out here, never imported from web/src/, so a change to the page
// that drops a word or adds a literal fails a test that says what it expected.
import { createHash } from 'node:crypto';
import { readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import type { Page } from '@playwright/test';
import { expect, test } from './app';
import { crawl, openDrawn } from './crawl';
import { buildState } from './state';

const PASS_WORDS: Record<string, string> = {
	not_started: 'Not started',
	waiting: 'Waiting',
	building: 'Building',
	stopped: 'Stopped',
	done: 'Done'
};
const ROUND_WORDS: Record<string, string> = {
	not_designed: 'Not designed yet',
	not_started: 'Not started',
	building: 'Building',
	done: 'Done'
};
const WIDGET_WORDS: Record<string, string> = {
	needs_you_now: 'Needs you now',
	ready_now: 'Ready now',
	scheduled: 'Today, about 8:00 pm',
	nothing_waiting: 'Nothing waiting'
};
/** Every pass and round code with its words. The two tables agree wherever a code is in both. */
const STATE_WORDS: Record<string, string> = { ...ROUND_WORDS, ...PASS_WORDS };

/**
 * Which builder in tests/state.ts shows which state. A state listed here that its builder does
 * not draw fails its test; a state in a table above that no builder lists fails the last test.
 */
const SHOWN_BY: Record<string, { pass: string[]; round: string[]; widget: string[]; asks: boolean }> = {
	started: {
		pass: ['not_started', 'building', 'done'],
		round: ['not_started', 'building', 'done'],
		widget: ['scheduled'],
		asks: true
	},
	two: { pass: ['waiting'], round: [], widget: ['ready_now'], asks: false },
	stopped: { pass: ['stopped'], round: [], widget: ['needs_you_now'], asks: false },
	redraw: { pass: [], round: ['not_designed'], widget: [], asks: false },
	big: { pass: [], round: [], widget: ['ready_now'], asks: false },
	quiet: { pass: [], round: [], widget: ['nothing_waiting'], asks: false }
};

interface Colored {
	where: string;
	code: string;
	text: string;
	asks: boolean;
}

/** Every colored state element on the drawn page: a bar segment or a `.state` span. */
async function colored(page: Page): Promise<Colored[]> {
	return page.evaluate((codes) => {
		const out: Colored[] = [];
		for (const el of document.querySelectorAll('.seg, .state')) {
			const code = [...el.classList].find((c) => codes.includes(c));
			if (!code) continue;
			const parent = el.closest('nav') ? 'nav' : 'main';
			out.push({
				where: `${parent} ${el.classList.contains('seg') ? '.seg' : '.state'}.${code}`,
				code,
				text: el.textContent ?? '',
				asks: el.classList.contains('asks')
			});
		}
		return out;
	}, Object.keys(STATE_WORDS));
}

/** One line per colored element whose text lacks its words, or an `asks` segment lacking its own. */
function missingWords(found: Colored[], address: string): string[] {
	const lines: string[] = [];
	for (const c of found) {
		if (!c.text.includes(STATE_WORDS[c.code])) {
			lines.push(`${address}: ${c.where} holds "${c.text}", not "${STATE_WORDS[c.code]}"`);
		}
		if (c.asks && !c.text.includes('Needs you')) {
			lines.push(`${address}: ${c.where} asks but holds "${c.text}", not "Needs you"`);
		}
	}
	return lines;
}

test.describe('words with color', () => {
	for (const [name, shown] of Object.entries(SHOWN_BY)) {
		test(`${name} shows its states with their words`, async ({ app, page }) => {
			test.slow();
			const clock = buildState(name, app.stateDir);
			await page.clock.setFixedTime(new Date(clock));
			await app.start();

			const problems: string[] = [];
			const segments = new Map<string, number>();
			const visit = async (address: string) => {
				const found = await colored(page);
				problems.push(...missingWords(found, address));
				if (address === '/') {
					for (const c of found) {
						if (c.where.includes('.seg.')) segments.set(c.code, (segments.get(c.code) ?? 0) + 1);
					}
				}
			};
			// The states that live on item pages are read there, so these builders walk every page.
			if (name === 'started' || name === 'stopped' || name === 'two') {
				await crawl(page, app.url, visit);
			} else {
				await openDrawn(page, app.url, '/');
				await visit('/');
			}
			expect(problems).toEqual([]);

			for (const code of [...shown.pass, ...shown.round]) {
				expect(segments.get(code) ?? 0, `${name} draws no .seg.${code} on /`).toBeGreaterThan(0);
			}
			if (shown.asks) {
				const asking = await page.locator('.seg.asks').count();
				expect(asking, `${name} draws no asking segment`).toBeGreaterThan(0);
			}

			// The widget holds its words in its own text, in the page's base ink, so no state color
			// can come to carry a state alone.
			for (const colorScheme of ['light', 'dark'] as const) {
				await page.emulateMedia({ colorScheme });
				const state = page.getByTestId('session-state');
				const words = shown.widget.map((code) => WIDGET_WORDS[code]);
				// A builder that lists no widget state is not read for its widget.
				if (words.length > 0) {
					await expect(state).toHaveText(words[0]);
					const ink = await state.evaluate((el) => getComputedStyle(el).color);
					const base = await page.evaluate(() => getComputedStyle(document.body).color);
					expect(ink, `the widget's color in ${colorScheme}`).toBe(base);
				}
			}
		});
	}

	test('every state in the tables is shown by a builder', () => {
		const passes = new Set(Object.values(SHOWN_BY).flatMap((s) => s.pass));
		const rounds = new Set(Object.values(SHOWN_BY).flatMap((s) => s.round));
		const widgets = new Set(Object.values(SHOWN_BY).flatMap((s) => s.widget));
		for (const code of Object.keys(PASS_WORDS)) expect(passes, `no builder shows pass ${code}`).toContain(code);
		for (const code of Object.keys(ROUND_WORDS)) expect(rounds, `no builder shows round ${code}`).toContain(code);
		for (const code of Object.keys(WIDGET_WORDS)) expect(widgets, `no builder shows ${code}`).toContain(code);
	});

	test('a banner holds its words in the text beside its color', async ({ app, page }) => {
		const clock = buildState('started', app.stateDir);
		await page.clock.setFixedTime(new Date(clock));
		await app.start();
		const banner = (kind: string) => page.locator(`.banner[data-banner="${kind}"] strong`);

		writeFileSync(path.join(app.stateDir, 'snapshot-error.json'), JSON.stringify({ message: 'releases/alpha.json: invalid character' }));
		await openDrawn(page, app.url, '/');
		await expect(banner('stale')).toHaveText('Stale');

		rmSync(path.join(app.stateDir, 'snapshot.json'));
		await expect(banner('no-state')).toHaveText('No state yet', { timeout: 15_000 });

		await app.stop();
		await expect(banner('not-connected')).toHaveText('Not connected', { timeout: 60_000 });
	});
});

/** Every file under `root` that `keep` accepts, as absolute paths. */
function filesUnder(root: string, keep: (rel: string) => boolean): string[] {
	const out: string[] = [];
	const walk = (dir: string) => {
		for (const entry of readdirSync(dir, { withFileTypes: true })) {
			const full = path.join(dir, entry.name);
			if (!keep(path.relative(root, full))) continue;
			if (entry.isDirectory()) walk(full);
			else out.push(full);
		}
	};
	walk(root);
	return out;
}

const webDir = path.resolve(import.meta.dirname, '..');

// The 148 CSS named colors, as Tailwind's own list holds them (transparent and currentcolor are
// not colors to pick, and the system colors are left out: `mark` is an ordinary word here).
const NAMED = [
	'aliceblue', 'antiquewhite', 'aqua', 'aquamarine', 'azure', 'beige', 'bisque', 'black',
	'blanchedalmond', 'blue', 'blueviolet', 'brown', 'burlywood', 'cadetblue', 'chartreuse',
	'chocolate', 'coral', 'cornflowerblue', 'cornsilk', 'crimson', 'cyan', 'darkblue', 'darkcyan',
	'darkgoldenrod', 'darkgray', 'darkgreen', 'darkgrey', 'darkkhaki', 'darkmagenta',
	'darkolivegreen', 'darkorange', 'darkorchid', 'darkred', 'darksalmon', 'darkseagreen',
	'darkslateblue', 'darkslategray', 'darkslategrey', 'darkturquoise', 'darkviolet', 'deeppink',
	'deepskyblue', 'dimgray', 'dimgrey', 'dodgerblue', 'firebrick', 'floralwhite', 'forestgreen',
	'fuchsia', 'gainsboro', 'ghostwhite', 'gold', 'goldenrod', 'gray', 'green', 'greenyellow',
	'grey', 'honeydew', 'hotpink', 'indianred', 'indigo', 'ivory', 'khaki', 'lavender',
	'lavenderblush', 'lawngreen', 'lemonchiffon', 'lightblue', 'lightcoral', 'lightcyan',
	'lightgoldenrodyellow', 'lightgray', 'lightgreen', 'lightgrey', 'lightpink', 'lightsalmon',
	'lightseagreen', 'lightskyblue', 'lightslategray', 'lightslategrey', 'lightsteelblue',
	'lightyellow', 'lime', 'limegreen', 'linen', 'magenta', 'maroon', 'mediumaquamarine',
	'mediumblue', 'mediumorchid', 'mediumpurple', 'mediumseagreen', 'mediumslateblue',
	'mediumspringgreen', 'mediumturquoise', 'mediumvioletred', 'midnightblue', 'mintcream',
	'mistyrose', 'moccasin', 'navajowhite', 'navy', 'oldlace', 'olive', 'olivedrab', 'orange',
	'orangered', 'orchid', 'palegoldenrod', 'palegreen', 'paleturquoise', 'palevioletred',
	'papayawhip', 'peachpuff', 'peru', 'pink', 'plum', 'powderblue', 'purple', 'rebeccapurple',
	'red', 'rosybrown', 'royalblue', 'saddlebrown', 'salmon', 'sandybrown', 'seagreen', 'seashell',
	'sienna', 'silver', 'skyblue', 'slateblue', 'slategray', 'slategrey', 'snow', 'springgreen',
	'steelblue', 'tan', 'teal', 'thistle', 'tomato', 'turquoise', 'violet', 'wheat', 'white',
	'whitesmoke', 'yellow', 'yellowgreen'
];
const PALETTES = [
	'slate', 'gray', 'zinc', 'neutral', 'stone', 'red', 'orange', 'amber', 'yellow', 'lime', 'green',
	'emerald', 'teal', 'cyan', 'sky', 'blue', 'indigo', 'violet', 'purple', 'fuchsia', 'pink', 'rose',
	'black', 'white'
];

const LITERALS: { what: string; pattern: RegExp }[] = [
	// The trailing boundary keeps Svelte's `{#each` from matching.
	{ what: 'a hex color', pattern: /#([0-9a-fA-F]{8}|[0-9a-fA-F]{6}|[0-9a-fA-F]{3,4})(?![\w-])/ },
	{
		what: 'a color function',
		pattern: /(?<![\w-])(rgb|rgba|hsl|hsla|hwb|lab|lch|oklab|oklch|color|color-mix|light-dark)\(/i
	},
	{ what: 'a named color', pattern: new RegExp(`(?<![\\w-])(${NAMED.join('|')})(?![\\w-])`, 'i') },
	{ what: 'an arbitrary Tailwind color', pattern: /\b(bg|text|border|ring|outline|fill|stroke|decoration|accent|caret|shadow|from|via|to)-\[(#|rgb|hsl|oklch|oklab|color|lab|lch|hwb)/i },
	{
		what: 'a Tailwind palette class',
		pattern: new RegExp(
			`(?<![\\w-])(bg|text|border|ring|outline|fill|stroke|decoration|accent|caret|shadow|from|via|to)-(${[...NAMED, ...PALETTES].join('|')})(-\\d+)?(?![\\w-])`
		)
	}
];

test.describe('Survey tokens only', () => {
	test('src holds no color literal', () => {
		const files = filesUnder(path.join(webDir, 'src'), () => true).filter((f) => /\.(css|svelte)$/.test(f));
		expect(files.length).toBeGreaterThan(10);
		const found: string[] = [];
		for (const file of files) {
			readFileSync(file, 'utf8')
				.split('\n')
				.forEach((line, i) => {
					for (const { what, pattern } of LITERALS) {
						const hit = pattern.exec(line);
						if (hit) found.push(`${path.relative(webDir, file)}:${i + 1}: ${what}, "${hit[0]}"`);
					}
				});
		}
		expect(found).toEqual([]);
	});

	test('the web folder holds no copy of Survey', () => {
		const surveyDir = path.join(webDir, 'node_modules', 'survey');
		const hash = (file: string) => createHash('sha256').update(readFileSync(file)).digest('hex');
		const survey = filesUnder(surveyDir, () => true);
		expect(survey.length, 'node_modules/survey is missing').toBeGreaterThan(3);
		const surveyHashes = new Map(survey.map((f) => [hash(f), path.relative(surveyDir, f)]));
		// A line of Survey's theme longer than this is a line nobody writes by chance.
		const LONG_LINE = 40;
		const surveyLines = new Set<string>();
		for (const f of survey.filter((f) => /theme\/[^/]+\.css$/.test(f))) {
			for (const line of readFileSync(f, 'utf8').split('\n')) {
				if (line.trim().length > LONG_LINE) surveyLines.add(line.trim());
			}
		}
		expect(surveyLines.size, 'Survey theme lines were not read').toBeGreaterThan(20);

		const skipped = new Set(['node_modules', 'build', '.svelte-kit', 'test-results']);
		const ours = filesUnder(webDir, (rel) => !skipped.has(rel.split(path.sep)[0]));
		expect(ours.length).toBeGreaterThan(30);
		const found: string[] = [];
		for (const file of ours) {
			const rel = path.relative(webDir, file);
			const same = surveyHashes.get(hash(file));
			if (same) found.push(`${rel}: a byte-for-byte copy of survey/${same}`);
			const lines = readFileSync(file, 'utf8').split('\n').map((l) => l.trim());
			const copied = lines.find((l) => surveyLines.has(l));
			if (copied) found.push(`${rel}: holds a line of Survey's theme, "${copied.slice(0, 60)}"`);
		}
		expect(found).toEqual([]);
	});
});
