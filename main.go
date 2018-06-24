package main

import (
	"fmt"
	"os"
	"time"

	"github.com/brandur/perpetual/updater"
)

func main() {
	_, err := updater.Update(nil, aeons, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during update: %v\n", err)
	}
}

//
// Private
//

// hundredYears represents a hundred years of time. This is used as a hack to
// get to ten thousand years, and is a convenient large block of time to add
// up which still fits within time.Duration's maximum size of ~290 years.
const hundredYears = time.Hour * 24 * 365 * 100

var aeons []*updater.Aeon

// time.Parse won't parse a 5-digit years, so we need a little hackiness to get
// us to ten thousand. Unfortunately, this is nowhere near as clean as other
// times because we know that leap years, etc. won't be handled well. This is
// probably made okay because it's unlikely that we ever get there.
func addTenThousandYears(t time.Time) time.Time {
	// 100 * 100 = 10,000
	for i := 0; i < 100; i++ {
		t.Add(hundredYears)
	}
	return t
}

func init() {
	aeons = []*updater.Aeon{
		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2018"), // base time
			Message: "Aeon 001 message"},

		{Target: updater.MustParseTime("Jun 25 08:00:00 PST 2018"), // 1 day
			Message: "Aeon 002 message"},

		{Target: updater.MustParseTime("Jul 01 08:00:00 PST 2018"), // 1 week
			Message: "Aeon 003 message"},

		{Target: updater.MustParseTime("Jul 24 08:00:00 PST 2018"), // 1 month
			Message: "Aeon 004 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2019"), // 1 year
			Message: "Aeon 005 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2028"), // 10 years
			Message: "Aeon 006 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 2118"), // 100 years
			Message: "Aeon 007 message"},

		{Target: updater.MustParseTime("Jun 24 08:00:00 PST 3018"), // 1,000 years
			Message: "Aeon 008 message"},

		{Target: addTenThousandYears(updater.MustParseTime("Jun 24 08:00:00 PST 2018")), // 10,000 years
			Message: "Aeon 009 message"},
	}
}
