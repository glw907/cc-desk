# Fixtures

`state/` is a cc-desk state folder; each test copies it into a temporary folder. Its `alpha` release is the worked case for passes 2 to 4: twelve finished passes at the clock 2026-11-20T23:00:00Z (2:00 pm Alaska time), a pace of 1.2 with a range of 0.8 to 1.5 from the last ten. `state/snapshot.json` is what the snapshot verb prints for it at that clock, and a test fails when it falls behind.

Its `beta` release has two finished passes (planned 40 and 50, taking 45 and 55), so its estimate is rough: pace 1.0 and no range. Round `b3` needs Geoff and has no design, and its predecessor pass `b-three` is planned at 480 minutes, so at the 2:00 pm clock `b3` starts at 10:00 pm, after alpha's 9:30 pm range end. Tests that need a written question for beta record `inputs/design-b3.json`.
