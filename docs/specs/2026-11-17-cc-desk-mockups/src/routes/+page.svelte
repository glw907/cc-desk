<script lang="ts">
	import ReleaseBoard from '../lib/ReleaseBoard.svelte';
	import { ahead, fmt, releases, waiting } from '../lib/mock.ts';
	const when = [
		{ href: '/r/cairn/7', project: 'cairn', round: 'round 7. Split the edit screen', what: '2 questions about how the edit screen is laid out.', due: 'Needed by Thursday, about 2 pm.' },
		{ href: '/r/cairn/8', project: 'cairn', round: 'round 8. Waymark and Pennant as themes', what: '1 question about which theme a new site starts with.', due: 'Needed by Friday, about 11 am, and asked early.' },
		{ href: '/r/dubplate/6', project: 'dubplate', round: 'round 6. The library screens', what: '4 questions about how the library looks and sorts.', due: 'Needed by tomorrow, about 10 am. This one sets the time.' }
	];
</script>

<svelte:head><title>Approval Queue</title></svelte:head>

<main class="desk">
	<section class="next grid gap-3">
		<h2>One thing is waiting for you</h2>
		{#each waiting as w (w.title)}
			<a class="front-row" href={w.href}>
				<span class="font-semibold">{w.title}</span>
				<span class="muted block text-sm">{w.project}. {w.note}. Sent {w.when}.</span>
			</a>
		{/each}
	</section>

	<section class="next grid gap-3">
		<h2>Your next decision session: tomorrow, about 9:30 am</h2>
		<p class="muted">Nothing running needs you before then. The estimate comes from how fast finished passes ran against their plans, and the likely range is 8:45 to 10:30 am.</p>
		<ul class="ahead">
			{#each when as w (w.round)}
				<li>
					<a href={w.href}><strong>{w.project}</strong>, {w.round}</a>
					<span class="muted">{w.what} {w.due}</span>
				</li>
			{/each}
		</ul>
		<p class="muted">Seven questions from three rounds, in two projects. After that session, nothing is expected to need you until Friday afternoon.</p>
	</section>

	<section class="grid gap-3">
		<h2>Two releases are under way</h2>
		<div class="boards">
			{#each releases as release (release.project)}
				<ReleaseBoard {release} />
			{/each}
		</div>
	</section>
</main>
