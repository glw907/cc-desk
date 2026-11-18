# Fixtures

`state/` is a cc-desk state folder; each test copies it into a temporary folder. Its `alpha` release is the worked case for passes 2 to 4: twelve finished passes at the clock 2026-11-20T23:00:00Z (2:00 pm Alaska time), a pace of 1.2 with a range of 0.8 to 1.5 from the last ten. `state/snapshot.json` is what the snapshot verb prints for it at that clock, and a test fails when it falls behind.

Its `beta` release has two finished passes (planned 40 and 50, taking 45 and 55), so its estimate is rough: pace 1.0 and no range. Round `b3` needs Geoff and has no design, and its predecessor pass `b-three` is planned at 480 minutes, so at the 2:00 pm clock `b3` starts at 10:00 pm, after alpha's 9:30 pm range end. Tests that need a written question for beta record `inputs/design-b3.json`.

`order/` is a second state folder for the start rules and the waiting order. It has two releases, `north` and `south`, each with one designed round that needs Geoff, and no `usage.json` or `settings.json`, so the usage stop is 100 and the first verb takes a reading. Every question is unanswered, with the options `yes` and `no` and `yes` recommended. `north` has `n-free` (60 minutes), `n-two` (30), `n-big` (400, rests on `north-q1`), `n-guess` (200, rests on `north-q2`), and `n-after` (50, after `n-guess`). `south` has `s-free` (45), `s-two` (30), and `s-guess` (150, rests on `south-q1`).

`usage40.sh` stands in for `claude -p "/usage"` and prints a 40 percent result. Every e2e run passes it as `--usage-command` and runs the binary with a `PATH` of `/usr/bin:/bin`, so no test reaches the real `claude`.
