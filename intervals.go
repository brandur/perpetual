package main

import (
	"github.com/brandur/perpetual/updater"
)

func init() {
	intervals = []*updater.Interval{
		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2018"), // base time
			Message: "Interval 001 message"},

		{Target: updater.MustParseTime("Jun 25 08:00:00 PST 2018"), // 1 day
			Message: "Interval 002 message"},

		{Target: updater.MustParseTime("Jul 01 08:00:00 PST 2018"), // 1 week
			Message: "Interval 003 message"},

		{Target: updater.MustParseTime("Jul 24 08:00:00 PST 2018"), // 1 month
			Message: "Interval 004 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2019"), // 1 year
			Message: "Interval 005 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2028"), // 10 years
			Message: "Interval 006 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2118"), // 100 years
			Message: "Interval 007 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 3018"), // 1,000 years
			Message: "Interval 008 message"},

		{Target: addTenThousandYears(updater.MustParseTime("Jun 24 08:00:00 PST 2018")), // 10,000 years
			Message: "Interval 009 message"},
	}
}
