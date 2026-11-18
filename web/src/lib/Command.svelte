<!--
@component
The terminal command that opens a desk on an item, as text, with a link that opens it in a terminal
and a button that copies it. The link is left out when the handler would refuse a name. A copy that
the browser refuses says nothing.
-->
<script lang="ts">
	import { terminalLink } from './link';

	interface Props {
		/** The project the command names. */
		project: string;
		/** The round the command names. A pass page passes its round's name. */
		round: string;
	}

	let { project, round }: Props = $props();

	let copied = $state(false);
	const command = $derived(`cc-desk ${project} ${round}`);
	const href = $derived(terminalLink(project, round));

	async function copy() {
		copied = false;
		try {
			await navigator.clipboard.writeText(command);
			copied = true;
		} catch {
			// A refused copy says nothing.
		}
	}
</script>

<h2>Command</h2>
<p><code>{command}</code></p>
<p class="command-actions">
	{#if href}
		<a class="term-link" {href}>Open in a terminal</a>
	{/if}
	<button type="button" class="btn btn-sm" onclick={copy}>Copy command</button>
	<span class="sr-only" role="status">{copied ? 'Copied' : ''}</span>
</p>
