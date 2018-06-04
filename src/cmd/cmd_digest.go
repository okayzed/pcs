package cmd

import (
	"flag"

	"github.com/logv/sybil/src/sybil"
)

func RunDigestCmdLine() {
	flag.Parse()

	flags := sybil.DefaultFlags()
	if *flags.TABLE == "" {
		flag.PrintDefaults()
		return
	}

	if *flags.PROFILE {
		profile := sybil.RUN_PROFILER()
		defer profile.Start().Stop()
	}

	t := sybil.GetTable(*flags.DIR, *flags.TABLE)
	if !t.LoadTableInfo() {
		sybil.Warn("Couldn't read table info, exiting early")
		return
	}
	t.DigestRecords(0, &sybil.DigestSpec{
		SkipOutliers:  *flags.SKIP_OUTLIERS,
		RecycleMemory: *flags.RECYCLE_MEM,
	})
}
