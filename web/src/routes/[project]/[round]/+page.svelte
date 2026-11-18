<!--
@component
A round's page at `/<project>/<round>`: its name, summary, state, figures, what it needs from
Geoff, its passes as links, and the command that opens it. It draws nothing until the first good
snapshot arrives, then the not-found view when the address names no round.
-->
<script lang="ts">
	import { page } from '$app/state';
	import Command from '#lib/Command.svelte';
	import Figures from '#lib/Figures.svelte';
	import NotFound from '#lib/NotFound.svelte';
	import Questions from '#lib/Questions.svelte';
	import { poll } from '#lib/poll.svelte.ts';
	import { findRoundAt } from '#lib/sums.ts';
	import { formatMinutes, formatWhen } from '#lib/time.ts';
	import { word } from '#lib/words.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);
	const found = $derived(
		snapshot ? findRoundAt(snapshot, page.params.project ?? '', page.params.round ?? '') : undefined
	);
	const minutes = (n: number | null) => (n === null ? null : formatMinutes(n));
	const when = (iso: string | null, about = false) =>
		iso === null ? null : formatWhen(iso, poll.now, about);
</script>

{#if snapshot && !found}
	<NotFound address={page.url.pathname} />
{:else if found}
	{@const { release, round } = found}
	<main class="page">
		<div class="page-main">
			<p class="doc-meta muted"><a href="/">{release.project} to {release.name}</a></p>
			<h1 class="doc-title">{round.name}</h1>

			<div class="prose-doc mt-6">
				<p>{round.summary}</p>
				<p class="state-line">
					<span class="state {round.state}">{word('roundState', round.state)}</span>
				</p>

				<h2>Figures</h2>
				<Figures
					figures={[
						{ label: 'Planned', value: minutes(round.planned_minutes) },
						{ label: 'Actual', value: minutes(round.actual_minutes) },
						{ label: 'Work before it', value: minutes(round.minutes_before) },
						{ label: 'Needed by', value: when(round.needed_time) },
						...(round.needs.kind === 'not_designed'
							? [{ label: 'Expected design', value: when(round.forecast_time, true) }]
							: [])
					]}
				/>

				<h2>What it needs from you</h2>
				{#if round.needs.kind === 'questions'}
					<Questions questions={round.needs.questions} />
				{:else if round.needs.kind === 'not_designed'}
					<p>
						{word('needsKind', 'not_designed')}.
						{#if round.forecast_time}Expected design: {when(round.forecast_time, true)}.{/if}
					</p>
				{:else}
					<p>Nothing. {round.needs.reason ?? 'No reason recorded.'}</p>
				{/if}

				<h2>Its passes</h2>
				{#if round.passes.length > 0}
					<ol>
						{#each round.passes as pass (pass.name)}
							<li>
								<a href="/{release.project}/{round.name}/{pass.name}">{pass.name}</a>
								<span class="state {pass.state}">{word('passState', pass.state)}</span>
							</li>
						{/each}
					</ol>
				{:else}
					<p class="muted">No passes are written yet.</p>
				{/if}

				<Command project={release.project} round={round.name} />
			</div>
		</div>
	</main>
{/if}
