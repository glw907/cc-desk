<!--
@component
A round's whole design, folded under a summary. `html` is markdown-it's output and nothing else:
the server renders the file with raw HTML off, images off, and only http and https links kept, and
this is the one place a page draws markup it did not build. Without `html` it shows the sentence
saying why there is no design to show.
-->
<script lang="ts">
	import type { DesignRead } from './server/design.ts';

	interface Props {
		/** What the server made of the design file. */
		design: DesignRead;
	}

	let { design }: Props = $props();
</script>

{#if 'html' in design}
	<details class="design">
		<summary>The full design</summary>
		<!-- eslint-disable-next-line svelte/no-at-html-tags -- markdown-it output, html off -->
		<div class="design-body">{@html design.html}</div>
	</details>
{:else}
	<p class="muted">{design.why}</p>
{/if}
