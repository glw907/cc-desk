# Fixtures

`state/` is a cc-desk state folder; each test copies it into a temporary folder. Its `alpha` release is the worked case for passes 2 to 4: twelve finished passes at the clock 2026-11-20T23:00:00Z (2:00 pm Alaska time), a pace of 1.2 with a range of 0.8 to 1.5 from the last ten. `state/snapshot.json` is what the snapshot verb prints for it at that clock, and a test fails when it falls behind.
