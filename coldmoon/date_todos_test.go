package coldmoon

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestGetNamedTimeZoneOffsetNanoseconds(t *testing.T) {
	t.Run("UTC", func(t *testing.T) {
		if got := GetNamedTimeZoneOffsetNanoseconds("UTC", 0); got != 0 {
			t.Fatalf("GetNamedTimeZoneOffsetNanoseconds(UTC, 0) = %d, want 0", got)
		}
	})

	t.Run("political time zone observes daylight saving time", func(t *testing.T) {
		if _, err := time.LoadLocation("America/New_York"); err != nil {
			t.Skipf("IANA time-zone data is unavailable: %v", err)
		}

		winter := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
		summer := time.Date(2024, time.July, 1, 12, 0, 0, 0, time.UTC)
		if got, want := GetNamedTimeZoneOffsetNanoseconds("America/New_York", float64(winter.UnixNano())), -5*int64(time.Hour); got != want {
			t.Fatalf("winter offset = %d, want %d", got, want)
		}
		if got, want := GetNamedTimeZoneOffsetNanoseconds("America/New_York", float64(summer.UnixNano())), -4*int64(time.Hour); got != want {
			t.Fatalf("summer offset = %d, want %d", got, want)
		}
	})
}

func TestDateAbstractStringOperations(t *testing.T) {
	if got := SystemTimeZoneIdentifier(); got != "UTC" {
		t.Fatalf("SystemTimeZoneIdentifier() = %q, want UTC", got)
	}

	if got := LocalTime(-123456789); got != -123456789 {
		t.Fatalf("LocalTime(-123456789) = %v, want identity in UTC-only mode", got)
	}

	if got := DateString(0); got != "Thu Jan 01 1970" {
		t.Fatalf("DateString(0) = %q, want %q", got, "Thu Jan 01 1970")
	}

	yearNine := JSNumber(time.Date(9, time.January, 2, 0, 0, 0, 0, time.UTC).UnixMilli())
	if got := DateString(yearNine); got != "Fri Jan 02 0009" {
		t.Fatalf("DateString(year 9) = %q, want %q", got, "Fri Jan 02 0009")
	}

	negativeYear := JSNumber(time.Date(-1, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli())
	if got := DateString(negativeYear); !strings.HasSuffix(got, " -0001") {
		t.Fatalf("DateString(year -1) = %q, want a -0001 year suffix", got)
	}

	timeValue := MakeTime(7, 5, 9, 999)
	if got := TimeString(timeValue); got != "07:05:09 GMT" {
		t.Fatalf("TimeString(07:05:09.999) = %q, want %q", got, "07:05:09 GMT")
	}
	if got := TimeZoneString(0); got != "+0000" {
		t.Fatalf("TimeZoneString(0) = %q, want +0000", got)
	}
}

func TestToDateStringUsesECMAScriptComposition(t *testing.T) {
	if got := ToDateString(0); got != "Thu Jan 01 1970 00:00:00 GMT+0000" {
		t.Fatalf("ToDateString(0) = %q, want %q", got, "Thu Jan 01 1970 00:00:00 GMT+0000")
	}
	if got := ToDateString(JSNumber(math.NaN())); got != InvalidDate {
		t.Fatalf("ToDateString(NaN) = %q, want %q", got, InvalidDate)
	}
}
