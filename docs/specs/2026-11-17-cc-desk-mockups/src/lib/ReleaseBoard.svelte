<!--
@component
One release on the front page: every round in order, with the running rounds opened to show
their passes.
-->
<script lang="ts">
	import ScaleBar from './ScaleBar.svelte';
	import { doneCount, fmt, leftMin, pct, words, type Release } from './mock.ts';
	let { release }: { release: Release } = $props();
</script>

<section class="board">
	<header class="board-head">
		<h3><strong>{release.project}</strong> to {release.name}</h3>
		<p class="muted">{doneCount(release.rounds)} of {release.rounds.length} rounds done. {pct(release.rounds)}% of the work, about {fmt(leftMin(release.rounds))} minutes left.</p>
		<ScaleBar parts={release.rounds} label="{doneCount(release.rounds)} of {release.rounds.length} rounds done" />
	</header>

	<ol class="rounds">
		{#each release.rounds as round, i (i)}
			<li class={round.state}>
				<span class="n">{i + 1}</span>
				<div>
					<p class="round-line">
						<a class="t" href="/r/{release.project}/{i + 1}">{round.title}</a>
						{#if round.asks && round.state === 'todo'}
							<span class="s needs"><span class="mark" aria-hidden="true"></span>{round.asks.count} {round.asks.count === 1 ? 'question' : 'questions'} for you</span>
						{:else}
							<span class="s">{words[round.state]}</span>
						{/if}
					</p>
					{#if round.passes}
						<ol class="passes">
							{#each round.passes as p, j (j)}
								<li class={p.state}>
									<span class="n">{j + 1}</span>
									<span class="t">{p.name}</span>
									<span class="s">{words[p.state]}</span>
								</li>
							{/each}
						</ol>
						{#if round.note}<p class="note">{round.note}</p>{/if}
					{/if}
				</div>
			</li>
		{/each}
	</ol>
</section>
