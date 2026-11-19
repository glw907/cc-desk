# Pricing a round's passes from their evidence

Three of the desk round's passes parked at planning, or nearly did, because their plans came out
over twice the round plan's estimate. Each wait for Geoff held up the passes after it. The
estimates and the paths the planners priced:

| Pass | Estimate | Plan path |
| --- | --- | --- |
| 1, the records | 120 | 209 |
| 2, the skills | 75 | 170 |
| 4, the page | 90 | 170 |
| 3, the desk | 100 | 255 |

The round plan priced each pass from its outline. It left out the work of proving each done
item: the live scripts and their reruns, new fixtures, a wording loop, the lines one pass leaves
to the next, and folding the plan review's findings. Every planner that priced clause by clause
came out about twice as high.

Two causes sit in the process, not the code:

1. The round planner prices a pass without the per-clause table its pass planner later writes.
   It should price each pass from its done items' evidence, live or Go test, the same way.
2. "Past twice the estimate, stop for Geoff" can't tell a plan whose scope grew from one that is
   priced in full. The desk round added standing answer 28 midway, after two such stops.
   A round plan should carry that answer from the start.

A third stop came from the order of the passes: pass 2's review escalated because its skills
describe a desk that pass 3 builds. A round plan that orders a pass ahead of what it describes
should say so, so the review reads it as expected.

cc-desk should carry the fix, since it runs every round from Move on. Two parts fit it:

1. **An estimate rule on the release.** A release could name a rule like standing answer 28,
   the way it names `ask_rules`: a plan that parks only over its estimate, with its scope
   matching the design, is raised and launched by the conductor and never reaches the desk. The
   desk's brief then lists each raise as something done, not a question.
2. **A miss signal after the first pass.** When a round's first plan path runs well past its
   estimate, the brief could say so and show what the later passes would come to at the same
   ratio. Geoff could then answer once for the whole round, before a later pass parks.

The Move round's design should settle both, and fold the evidence pricing into its own round
plan. The process half belongs in `pass-core`'s `references/round.md` and
`references/planner.md`, a workstation change through `claude-infra`.
