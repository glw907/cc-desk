# Record a real hook input and check its session_id

The prompt hook's tests read `e2e/testdata/prompts/hook-bash.json`, the example in the hooks
reference. No live session has fed `cc-desk prompt hook` yet, so two things rest on the
documentation alone.

1. **The `session_id` match.** The hook finds its record by comparing the input's `session_id`
   with the `session_id` the launch stored from the job's `state.json`. If a real input carries a
   different id for the same session, every prompt lands under `hook-unmatched/` and nothing shows
   as "Needs you now". In pass 5's live run, raise one prompt and compare the two ids.
2. **The input's shape.** Save one real PermissionRequest input beside the documented one, and add
   a case to `parseHookPrompt`'s table test and to `TestPromptHook` that reads it.
