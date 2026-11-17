<script lang="ts">
	import { page } from '$app/state';
	import { before, fmt, releases, words } from '../../../../lib/mock.ts';
	const release = $derived(releases.find((r) => r.project === page.params.project)!);
	const n = $derived(Number(page.params.n));
	const round = $derived(release.rounds[n - 1]);
	const wait = $derived(before(release, n - 1));
</script>

<svelte:head><title>{round.title}</title></svelte:head>

<main class="page">
	<div class="page-main">
		<p class="doc-meta muted"><a href="/">{release.project} to {release.name}</a>. Round {n} of {release.rounds.length}.</p>
		<h1 class="doc-title">{round.title}</h1>

		<div class="prose-doc mt-6">
			{#if round.state === 'todo'}
				<p class="muted">Not started. It begins after about {fmt(wait)} minutes of work, and is planned at {fmt(round.min)} minutes.</p>
			{:else}
				<p class="muted">{words[round.state]}. Planned at {fmt(round.min)} minutes.</p>
			{/if}

			<p>{round.about ?? 'This round has a place in the plan and no written design yet. Its design is written while the rounds before it build.'}</p>

			<h2>What it needs from you</h2>
			{#if round.asks}
				<p>{round.asks.count} {round.asks.count === 1 ? 'question' : 'questions'}, about {round.asks.about}.</p>
				{#if round.asks.questions}
					<ul class="calls">
						{#each round.asks.questions as q (q)}<li>{q}</li>{/each}
					</ul>
					<p>Both are ready to answer now. Answered early, this round starts and finishes without stopping for you.</p>
				{:else}
					<p>The questions are not written yet.</p>
				{/if}
			{:else}
				<p>Nothing. No part of this round changes what a person sees or types, and it fits the spend you approved. You read its report when it ends.</p>
			{/if}

			{#if round.passes || round.planned}
				<h2>Its passes</h2>
				<ol>
					{#each round.passes ?? [] as p (p.name)}<li>{p.name}{words[p.state] ? `. ${words[p.state]}.` : ''}</li>{/each}
					{#each round.planned ?? [] as p (p)}<li>{p}</li>{/each}
				</ol>
			{/if}
		</div>

		{#if round.asks?.questions}
			<section class="decide">
				<button class="btn btn-primary">Answer these now</button>
				<p class="muted mt-3">Or leave them for the next sitting, with the questions from three other rounds.</p>
			</section>
		{/if}
	</div>
</main>
