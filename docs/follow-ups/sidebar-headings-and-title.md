# The sidebar's section names don't read as headings

Geoff, 2026-11-20, after his sitting: "The sidebar has the following sections: Ruling, Waiting
for you, and Running. But these don't really look like headings or section headers. The sidebar
probably needs a bit of design refinement to address that." And: "the cc-desk title should be a
bit more prominent."

The sidebar is `web/src/lib/Sidebar.svelte`, inside the rail in `web/src/routes/+layout.svelte`.
"Waiting for you", "Reports to read", and "Running" are `h2.rail-heading` elements. "Rulings" is
a `rail-row` link set beside them, so it reads as one more name in the same list. The title is
the `wordmark` link at `text-xl`.

Geoff asked for "a gentle sidebar session as part of the 1.1 brainstorm". The 1.1 design prompt
names this file, and the design session runs `standards:design-refinement` on the live page for
the sidebar headings and the title, with Geoff looking at each probe in Chromium.
