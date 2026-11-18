<!--
@component
One release on the front page: every round in order, with a building round, or one holding a
waiting or stopped pass, opened to show its passes. A waiting pass shows why it waits.
-->
<script lang="ts">
	import ScaleBar from './ScaleBar.svelte';
	import type { Release, Snapshot } from './snapshot';
	import { count, doneCount, isOpen, needsGeoff, percentDone, roundParts } from './sums';
	import { word } from './words';

	interface Props {
		/** The release whose rounds are listed, with the running ones opened to their passes. */
		release: Release;
		/** Where a waiting pass's limits and reset text come from. */
		snapshot: Snapshot;
	}

	let { release, snapshot }: Props = $props();

	const roundsDone = $derived(
		`${doneCount(release.rounds)} of ${release.rounds.length} rounds done`
	);

	function waitingFor(pass: string) {
		return snapshot.waiting.find((w) => w.project === release.project && w.pass === pass);
	}
</script>

<section class="board" data-release="{release.project}/{release.name}">
	<header class="board-head">
		<h3>
			<strong>{release.project}</strong> to {release.name}
			{#if release.rough}<span class="rough">Rough</span>{/if}
		</h3>
		<p class="muted">{roundsDone}. {percentDone(release.rounds)}% of the work.</p>
		<ScaleBar parts={roundParts(release.rounds)} label={roundsDone} />
	</header>

	<ol class="rounds">
		{#each release.rounds as round (round.name)}
			<li class={round.state} data-round="{release.project}/{round.name}">
				<p class="round-line">
					<a class="t" href="/{release.project}/{round.name}">{round.name}</a>
					{#if needsGeoff(round) && round.needs.kind === 'questions'}
						<span class="s needs"
							>{count(round.needs.questions.filter((q) => !q.answered).length, 'question')} for you</span
						>
					{:else}
						<span class="s">{word('roundState', round.state)}</span>
					{/if}
				</p>
				{#if isOpen(round)}
					<ol class="passes">
						{#each round.passes as pass (pass.name)}
							{@const wait = waitingFor(pass.name)}
							<li class={pass.state} data-pass="{release.project}/{pass.name}">
								<span class="t">{pass.name}</span>
								<span class="s state {pass.state}">{word('passState', pass.state)}</span>
								{#if pass.state === 'stopped' && pass.stop_reason}
									<span class="why">{word('stopReason', pass.stop_reason)}</span>
								{/if}
								{#if pass.state === 'waiting' && wait}
									<span class="why">
										{wait.limits.map((l) => word('limit', l)).join(', ')}{#if wait.reset_text}. {wait.reset_text}{/if}
									</span>
								{/if}
							</li>
						{/each}
					</ol>
				{/if}
			</li>
		{/each}
	</ol>
</section>
