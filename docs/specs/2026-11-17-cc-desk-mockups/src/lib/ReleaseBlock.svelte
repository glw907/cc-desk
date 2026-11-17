<!--
@component
One release in the sidebar: a bar of its rounds, then each running round hung beneath it on a
rule, with a bar of that round's passes.
-->
<script lang="ts">
	import ScaleBar from './ScaleBar.svelte';
	import { doneCount, nextAsk, pct, running, words, type Release } from './mock.ts';
	let { release }: { release: Release } = $props();
	const tone = { stalled: 'error', now: 'primary', review: 'primary', done: 'primary', todo: 'primary' } as const;
</script>

<article class="release">
	<p class="release-head">
		<span><strong>{release.project}</strong> to {release.name}</span>
	</p>
	<ScaleBar parts={release.rounds} label="{doneCount(release.rounds)} of {release.rounds.length} rounds done" />
	<p class="figures">
		<span>{doneCount(release.rounds)} of {release.rounds.length} rounds done</span>
		<span>{pct(release.rounds)}% of the work</span>
	</p>

	{#if nextAsk(release)}
		{@const a = nextAsk(release)!}
		<p class="figures asks-line"><a href="/r/{release.project}/{a.n}"><span class="mark" aria-hidden="true"></span>Needs you next at round {a.n}</a></p>
	{/if}

	<ul class="branch">
		{#each running(release) as { round, n } (n)}
			{@const passes = round.passes ?? []}
			<li>
				<p class="branch-head">
					<a class="branch-title" href="/r/{release.project}/{n}">Round {n}. {round.title}</a>
				</p>
				<ScaleBar parts={passes} label="{doneCount(passes)} of {passes.length} passes done" slim />
				<p class="figures">
					<span>{doneCount(passes)} of {passes.length} passes done</span>
					<span class="round-state tone-{tone[round.state]}"><span class="status" aria-hidden="true"></span>{words[round.state]}</span>
				</p>
			</li>
		{/each}
	</ul>
</article>
