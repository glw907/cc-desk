# A desk that lives through a clock step fails the start-time test

The ruling guard matches a desk by its process id and its start time, within 2 seconds. It builds
the start time from the boot time (`btime` in `/proc/stat`) plus the process's start ticks. Setting
the wall clock moves `btime`, so a desk that is already running when the clock steps by more than 2
seconds no longer matches its entry in `desks.json`. Test 2 then refuses a ruling from that desk.

The Desk round can fix this by storing the raw `starttime` ticks from `/proc/<pid>/stat` along with
the boot id from `/proc/sys/kernel/random/boot_id`. Ticks count from boot and don't move when the
clock is stepped. The window stays at 2 seconds until then.

The code is `Entry` in `internal/desk/desk.go` and `SameStart` in `internal/state/ruling.go`.
