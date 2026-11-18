<!--
@component
The rulings page at `/rulings`: every ruling Geoff has made, newest first, with the question it
answered (the en dash when it answered none) and his words. It draws nothing until the first good
snapshot arrives.
-->
<script lang="ts">
	import { poll } from '#lib/poll.svelte.ts';
	import { questionText } from '#lib/sums.ts';
	import { DASH, NONE, formatWhen } from '#lib/time.ts';

	const snapshot = $derived(poll.answer?.snapshot ?? null);
	const rulings = $derived(
		snapshot ? [...snapshot.rulings].sort((a, b) => Date.parse(b.time) - Date.parse(a.time)) : []
	);

	function question(project: string | undefined, id: string | undefined) {
		if (!snapshot || !project || !id) return null;
		return questionText(snapshot, project, id) ?? null;
	}
</script>

{#if snapshot}
	<main class="page">
		<div class="page-main">
			<p class="doc-meta muted"><a href="/">cc-desk</a></p>
			<h1 class="doc-title">Rulings</h1>
			<div class="prose-doc mt-6">
				{#if rulings.length === 0}
					<p>No rulings yet</p>
				{:else}
					<table class="table">
						<thead>
							<tr><th>When</th><th>Question</th><th>Ruling</th></tr>
						</thead>
						<tbody>
							{#each rulings as ruling (ruling.id)}
								{@const text = question(ruling.project, ruling.question)}
								<tr>
									<td>{formatWhen(ruling.time, poll.now)}</td>
									<td>
										{#if text}{text}{:else}<span aria-hidden="true">{DASH}</span><span class="sr-only">{NONE}</span>{/if}
									</td>
									<td>{ruling.words}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{/if}
			</div>
		</div>
	</main>
{/if}
