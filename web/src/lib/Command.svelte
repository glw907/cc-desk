<!--
@component
The terminal command that opens a desk on an item, shown as text for the reader to copy.
-->
<script lang="ts">
	interface Props {
		/** The project the command names. */
		project: string;
		/** The round the command names. A pass page passes its round's name. */
		round: string;
	}

	let { project, round }: Props = $props();
</script>

<h2>Command</h2>
<p><code>cc-desk {project} {round}</code></p>
