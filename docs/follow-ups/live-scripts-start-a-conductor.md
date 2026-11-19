# A live script's release can start a conductor in another script's run

The check now starts a conductor when a release's report falls due. The live scripts share one
scratch project and one state folder with the timer. `scripts/live-check.sh` merges both passes of
its release (`timer-live-a` and `timer-live-b`), so the release's report falls due and the next
check starts a conductor in the scratch project. If another script's run is live at that moment, the
conductor belongs to neither run, and no script ends it.

`scripts/live-repairs.sh` avoids this. Each run's round holds a pass that only waits, so no report
comes due in the scratch project.

Have `live-check.sh` add a held pass to its release as `live-repairs.sh` does, or have it end a
conductor it finds on the release it added before it prints its result.
