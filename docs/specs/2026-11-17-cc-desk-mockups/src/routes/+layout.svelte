<script lang="ts">
	import '../lib/survey/app.css';
	import '../lib/mock.css';
	import { page } from '$app/state';
	import ReleaseBlock from '../lib/ReleaseBlock.svelte';
	import { releases, reports, waiting } from '../lib/mock.ts';
	let { children } = $props();
	const here = (href: string) => (page.url.pathname === href ? 'page' : undefined);
</script>

<div class="drawer drawer-open">
	<input id="rail" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content">{@render children()}</div>
	<div class="drawer-side">
		<nav class="rail min-h-full p-4" style="width: var(--rail-width)" aria-label="Queue">
			<a href="/" class="wordmark block px-3 pb-6 pt-2 text-xl"><span class="site-benchmark"></span>Approval Queue</a>

			<a class="session" href="/">
				<span class="session-label">Next decision session</span>
				<span class="session-time">Tomorrow, about 9:30 am</span>
				<span class="session-sub">Likely between 8:45 and 10:30 am</span>
				<span class="session-sub">7 questions from 3 rounds, in cairn and dubplate</span>
			</a>

			<section class="rail-section" style="margin-top: var(--spacing-24px)">
				<h2 class="rail-heading">Waiting for you</h2>
				{#each waiting as w (w.title)}
					<a class="rail-row" href={w.href} aria-current={here(w.href)}>
						{w.title}
						<span class="sub"><span>{w.note}</span><span>{w.when}</span></span>
					</a>
				{/each}
			</section>

			<section class="rail-section">
				<h2 class="rail-heading">Reports to read</h2>
				{#each reports as r, i (r.title)}
					<a class="rail-row" href={r.href} aria-current={i === 0 ? here(r.href) : undefined}>
						{r.title}
						<span class="sub"><span>{r.note}</span><span>{r.when}</span></span>
					</a>
				{/each}
			</section>

			<section class="rail-section">
				<h2 class="rail-heading">Running</h2>
				{#each releases as release (release.project)}
					<ReleaseBlock {release} />
				{/each}
			</section>
		</nav>
	</div>
</div>
<p class="sketch-flag">Sketch. The names are real, the figures are made up.</p>
