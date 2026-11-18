# A project named `api` with a round named `state` has no page

The server answers `/api/state` with the snapshot, and a static route wins over `/[project]/[round]`.
A project named `api` whose round is named `state` would have a page no address can reach. No
fixture has one and no project is named `api` today. If one is ever registered, move the state
endpoint to an address a project name can't take, such as `/_state`, and update `web/src/lib/poll.svelte.ts`
and `tests/server.spec.ts` to match.
