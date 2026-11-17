// Sample data for the sketch. The names are real; the figures are made up.
export type State = 'done' | 'now' | 'review' | 'stalled' | 'todo';
export type Pass = { name: string; min: number; state: State };
export type Ask = { count: number; about: string; questions?: string[] };
export type Round = { title: string; min: number; state: State; passes?: Pass[]; note?: string; asks?: Ask; about?: string; planned?: string[] };
export type Release = { project: string; name: string; rounds: Round[] };

const cairn: [string, number][] = [
	['Delete the Go tool', 460],
	['One list of checks, one runner', 600],
	['Remove the published names you ruled out', 860],
	['Put every folder in the layer order', 740],
	['Typed errors everywhere', 860],
	['Stock pages in the package', 1031],
	['Split the edit screen', 980],
	['Waymark and Pennant as themes', 758],
	['Describe a made site in code', 830],
	["The command's frame and cairn doctor", 678],
	['Page commands and cairn update', 738],
	['Features, cairn add, events', 630],
	['Search, contact form, links', 580],
	['Credentials, cairn health, terminal screens', 1316],
	['cairn new, launcher, one live run', 1207],
	['Survey', 704],
	['Docs and release candidate', 1160]
];

const cairnExtra: Record<number, Partial<Round>> = {
	6: {
		asks: {
			count: 2,
			about: 'how the edit screen is laid out',
			questions: ['Where the page settings sit: a side panel, or a second tab?', 'Whether the preview opens beside the editor or replaces it.']
		},
		about: 'The edit screen is one 1,900-line file that draws the editor, the page settings, and the preview. This round splits it into three parts that load on their own, so the editor opens faster and each part can change without the others.',
		planned: ['Pull the page settings out of the editor', 'Pull the preview out of the editor', 'Load each part only when it opens', 'Bring the tests and the docs in line']
	},
	7: { asks: { count: 1, about: 'which theme a new site starts with' } },
	9: { asks: { count: 2, about: 'the names of two commands a person types' } },
	14: { asks: { count: 3, about: 'what cairn new asks a person, and in what order' } }
};

export const releases: Release[] = [
	{
		project: 'cairn',
		name: 'the first public release',
		rounds: cairn.map(([title, min], i) => ({
			title,
			min,
			state: i < 2 ? 'done' : i === 2 ? 'now' : 'todo',
			...cairnExtra[i],
			passes:
				i === 2
					? [
							{ name: 'Drop the three old command names', min: 180, state: 'done' },
							{ name: 'Remove the theme install commands', min: 260, state: 'now' },
							{ name: 'Remove the old import paths', min: 220, state: 'todo' },
							{ name: 'Bring the docs and demo sites in line', min: 200, state: 'todo' }
						]
					: undefined
		}))
	},
	{
		project: 'dubplate',
		name: 'the first working server',
		rounds: [
			{ title: 'The record store', min: 900, state: 'done' },
			{ title: 'Reading tags', min: 700, state: 'done' },
			{
				title: 'One engine for every kind of record',
				min: 1040,
				state: 'review',
				passes: [
					{ name: 'Move the id rule to one place', min: 240, state: 'done' },
					{ name: 'Give each table an owner', min: 420, state: 'done' },
					{ name: 'Split the server command by domain', min: 380, state: 'review' }
				]
			},
			{
				title: 'The catalog reads from the new engine',
				min: 900,
				state: 'stalled',
				note: 'Quiet for 25 minutes. The desk is restarting pass 2.',
				passes: [
					{ name: 'Catalog queries on the engine', min: 300, state: 'done' },
					{ name: 'Search and sort', min: 340, state: 'stalled' },
					{ name: 'Retire the old catalog code', min: 260, state: 'todo' }
				]
			},
			{ title: 'Intake', min: 2100, state: 'todo', asks: { count: 2, about: 'where an unmatched folder goes, and when to start' } },
			{ title: 'The library screens', min: 1400, state: 'todo', asks: { count: 4, about: 'how the library looks and sorts' } },
			{ title: 'Playback', min: 1100, state: 'todo' }
		]
	}
];

export const waiting = [
	{ href: '/d/design', project: 'dubplate', title: 'Approve the dubplate intake design?', note: '2 questions', when: '20 minutes ago' }
];

export const reports = [
	{ href: '/d/report', project: 'cairn', title: 'cairn has one list of checks and one runner', note: '3 judgment calls, 2 planning misses', when: '1 hour ago' },
	{ href: '/d/report', project: 'cairn', title: 'The Go tool is deleted', note: '1 judgment call', when: 'this morning' }
];

export const words: Record<State, string> = { done: 'Done', now: 'Building', review: 'In review', stalled: 'Stalled', todo: '' };
const sum = (xs: { min: number }[]) => xs.reduce((n, x) => n + x.min, 0);
export const doneCount = (xs: { state: State }[]) => xs.filter((x) => x.state === 'done').length;
export const pct = (xs: { min: number; state: State }[]) => Math.round((100 * sum(xs.filter((x) => x.state === 'done'))) / sum(xs));
export const leftMin = (xs: { min: number; state: State }[]) => sum(xs.filter((x) => x.state !== 'done'));
export const fmt = (n: number) => n.toLocaleString('en-US');
export const running = (r: Release) => r.rounds.map((round, i) => ({ round, n: i + 1 })).filter((x) => x.round.passes);

/** The work, in minutes, that stands between now and the start of round `i`. */
export const before = (r: Release, i: number) => r.rounds.slice(0, i).filter((x) => x.state !== 'done').reduce((n, x) => n + x.min, 0);
export const ahead = releases.flatMap((r) => r.rounds.map((round, i) => ({ release: r, round, n: i + 1, wait: before(r, i) })).filter((x) => x.round.asks && x.round.state === 'todo'));
export const nextAsk = (r: Release) => ahead.find((x) => x.release === r);
