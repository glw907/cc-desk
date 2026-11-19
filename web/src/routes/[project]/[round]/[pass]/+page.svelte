<!--
@component
A pass's page at `/<project>/<round>/<pass>`: its name, summary, state, figures, what it needs,
and the command of its round. A stopped pass shows its reason and when the stop began. A pass
with no questions says "Nothing." with its own reason (the snapshot's `needs_reason`). The
permission prompts and refused commands of the sessions on the pass are listed with their times.
-->
<script lang="ts">
	import { page } from '$app/state';
	import Command from '#lib/Command.svelte';
	import Figures from '#lib/Figures.svelte';
	import NotFound from '#lib/NotFound.svelte';
	import Questions from '#lib/Questions.svelte';
	import { poll } from '#lib/poll.svelte.ts';
	import { findPassAt } from '#lib/sums.ts';
	import { formatMinutes, formatWhen } from '#lib/time.ts';
	import { word } from '#lib/words.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);
	const found = $derived(
		snapshot
			? findPassAt(
					snapshot,
					page.params.project ?? '',
					page.params.round ?? '',
					page.params.pass ?? ''
				)
			: undefined
	);
	const sessions = $derived(
		(snapshot?.sessions ?? []).filter(
			(s) =>
				s.project === page.params.project &&
				s.round === page.params.round &&
				s.pass === page.params.pass
		)
	);
	const prompts = $derived(sessions.flatMap((s) => s.prompts));
	const refusals = $derived(sessions.flatMap((s) => s.refusals));
	const minutes = (n: number | null) => (n === null ? null : formatMinutes(n));
</script>

{#if snapshot && !found}
	<NotFound address={page.url.pathname} />
{:else if found}
	{@const { release, round, pass } = found}
	<main class="page">
		<div class="page-main">
			<p class="doc-meta muted">
				<a href="/{release.project}/{round.name}">{release.project}/{round.name}</a>
			</p>
			<h1 class="doc-title">{pass.name}</h1>

			<div class="prose-doc mt-6">
				<p>{pass.summary}</p>
				<p class="state-line">
					<span class="state {pass.state}">{word('passState', pass.state)}</span>
					{#if pass.state === 'stopped' && pass.stop_reason}
						<span class="why">
							{word('stopReason', pass.stop_reason)}{#if pass.stopped_at}, {formatWhen(pass.stopped_at, poll.now)}{/if}
						</span>
					{/if}
				</p>

				<h2>Figures</h2>
				<Figures
					figures={[
						{ label: 'Planned', value: minutes(pass.planned_minutes) },
						{ label: 'Actual', value: minutes(pass.actual_minutes) },
						{ label: 'Work before it', value: minutes(pass.minutes_before) },
						{
							label: 'Needed by',
							value: pass.needed_time === null ? null : formatWhen(pass.needed_time, poll.now)
						}
					]}
				/>

				<h2>What it needs from you</h2>
				{#if pass.questions.length > 0}
					<Questions questions={pass.questions} />
				{:else}
					<p>Nothing. {pass.needs_reason ?? 'No reason recorded.'}</p>
				{/if}

				<h2>Permission prompts</h2>
				{#if prompts.length > 0}
					<ul>
						{#each prompts as prompt (prompt.time + (prompt.command ?? ''))}
							<li>
								{formatWhen(prompt.time, poll.now)}: {prompt.tool}{#if prompt.command}, <code>{prompt.command}</code>{/if}.
								{#if prompt.suggested_rule}Rule it would add: <code>{prompt.suggested_rule}</code>.{/if}
								{prompt.answer ? `Answered: ${prompt.answer}.` : 'Not answered yet.'}
							</li>
						{/each}
					</ul>
				{:else}
					<p class="muted">No session on this pass has stopped on a permission prompt.</p>
				{/if}

				<h2>Refused commands</h2>
				{#if refusals.length > 0}
					<ul>
						{#each refusals as refusal (refusal.time + refusal.target)}
							<li>
								{formatWhen(refusal.time, poll.now)}: {refusal.tool}, <code>{refusal.target}</code>. {refusal.reason}.
							</li>
						{/each}
					</ul>
				{:else}
					<p class="muted">No command on this pass was refused.</p>
				{/if}

				<Command project={release.project} round={round.name} />
			</div>
		</div>
	</main>
{/if}
