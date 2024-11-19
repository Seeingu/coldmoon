package coldmoon

import (
	"math"
	"time"
)

type DateObject struct {
	*Object
	Data float64
}

func NewDatePrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype)
	agent := realm.Agent

	var valueOf BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if o, ok := ValueGetObject(this); ok {
			if date, ok := o.(*DateObject); ok {
				return NewNumberValue(date.Data)
			} else {
				panic("TypeError")
			}
		} else {
			panic("TypeError")
		}
	}
	var toPrimitive BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		hintValue := args[0]
		if !ValueIsObject(this) {
			panic("TypeError")
		}
		o := MustGetObject(this)
		if !ValueIs[*StringValue](hintValue) {
			panic("TypeError")
		}
		hint := hintValue.(*StringValue).Data
		var tryFirst PreferredType
		if hint == "string" || hint == "default" {
			tryFirst = PreferredTypeString
		} else if hint == "number" {
			tryFirst = PreferredTypeNumber
		} else {
			panic("TypeError")
		}
		return o.OrdinaryToPrimitive(tryFirst)
	}
	var toString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		date := RequireInternalSlot[*DateObject](this)
		return NewStringValue(ToDateString(date.Data))
	}
	var toISOString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if !math.IsInf(tv, 0) {
			panic("RangeError")
		}

		// Use go standard formatter
		//_ = YearFromTime(tv)
		//_ = MonthFromTime(tv)
		//_= DateFromTime(tv)
		//_= HourFromTime(tv)
		//_= MinFromTime(tv)
		//_ = SecFromTime(tv)
		//_= msFromTime(tv)

		t, err := time.Parse(time.RFC3339, time.Unix(int64(tv), 0).Format(time.RFC3339))
		if err != nil {
			panic("RangeError")
		}
		return NewStringValue(
			t.Format("2006-01-02T15:04:05.999Z"),
		)
	}
	var toJSON BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		o := MustGetObject(this)
		tv := ToPrimitive(agent, NewValueFromObject(o), PreferredTypeNumber)
		if n, ok := ValueGet[*NumberValue](tv); ok && !n.IsFinite() {
			return NullValue
		}
		return ValueInvoke(agent, NewValueFromObject(o), NewStringPropertyKey("toISOString"), nil)
	}
	var toUTCString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		t, err := time.Parse(time.RFC3339, time.Unix(int64(tv), 0).Format(time.RFC3339))
		if err != nil {
			panic("RangeError")
		}
		return NewStringValue(t.Format(time.RFC1123))
	}
	var toDateString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := MustGetObject(this).(*DateObject)
		tv := dateObject.Data
		if math.IsNaN(tv) {
			return NewStringValue("Invalid Date")
		}
		return NewStringValue(DateString(LocalTime(tv)))
	}
	var toTimeString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := MustGetObject(this).(*DateObject)
		tv := dateObject.Data
		if math.IsNaN(tv) {
			return NewStringValue("Invalid Date")
		}
		return NewStringValue(TimeString(LocalTime(tv)) + TimeZoneString(tv))
	}
	var toLocaleString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		return ValueInvoke(agent, this, NewStringPropertyKey("toString"), nil)
	}
	var toLocaleDateString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		return ValueInvoke(agent, this, NewStringPropertyKey("toDateString"), nil)
	}
	var toLocaleTimeString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		return ValueInvoke(agent, this, NewStringPropertyKey("toTimeString"), nil)
	}
	var getTimezoneOffset BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := MustGetObject(this).(*DateObject)
		tv := dateObject.Data
		if math.IsNaN(tv) {
			return NewNumberValue(math.NaN())
		}
		return NewNumberValue(tv - LocalTime(tv)/MS_PER_MIN)
	}

	DefineBuiltinFunction(object, "valueOf", valueOf, 0, realm)
	DefineBuiltinFunction(object, "toString", toString, 0, realm)
	DefineBuiltinFunction(object, "toISOString", toISOString, 0, realm)
	DefineBuiltinFunction(object, "toJSON", toJSON, 1, realm)
	DefineBuiltinFunction(object, "toUTCString", toUTCString, 0, realm)
	DefineBuiltinFunction(object, "toDateString", toDateString, 0, realm)
	DefineBuiltinFunction(object, "toTimeString", toTimeString, 0, realm)
	DefineBuiltinFunction(object, "toLocaleString", toLocaleString, 0, realm)
	DefineBuiltinFunction(object, "toLocaleDateString", toLocaleDateString, 0, realm)
	DefineBuiltinFunction(object, "toLocaleTimeString", toLocaleTimeString, 0, realm)
	DefineBuiltinFunction(object, "getTimezoneOffset", getTimezoneOffset, 0, realm)

	DefineBuiltinFunctionWithAttributes(object, "@@toPrimitive", toPrimitive, 1, realm, PropertyDescriptorAttributes{
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}

const MS_PER_DAY = 86400000
const MS_PER_MIN = 60000

// 21.4.1.3
func Day(t float64) float64 {
	// FIXME: use time package
	msPerDay := MS_PER_DAY
	return t / float64(msPerDay)
}

func DaysInYear(y float64) float64 {
	if math.Mod(y, 4) != 0 {
		return 365
	}
	if math.Mod(y, 100) != 0 {
		return 366
	}
	if math.Mod(y, 400) != 0 {
		return 365
	}
	return 366
}

func DayFromYear(y float64) float64 {
	return 365*(y-1970) + math.Floor((y-1969)/4) - math.Floor((y-1901)/100) + math.Floor((y-1601)/400)
}

func TimeFromYear(y float64) float64 {
	return MS_PER_DAY * DayFromYear(y)
}

func YearFromTime(t float64) float64 {
	year := t / ((365.2425 * MS_PER_DAY) + 1970)
	t2 := TimeFromYear(year)
	if t2 > t {
		return year - 1
	}
	if t2+(DaysInYear(year)*MS_PER_DAY) <= t {
		return year + 1
	}
	return year
}

func DayWithinYear(t float64) float64 {
	return Day(t) - DayFromYear(YearFromTime(t))
}

func InLeapYear(t float64) bool {
	if DaysInYear(YearFromTime(t)) == 366 {
		return true
	}
	return false
}

func MonthFromTime(t float64) float64 {
	day := DayWithinYear(t)
	if InLeapYear(t) {
		if day < 31 {
			return 0
		}
		if day < 60 {
			return 1
		}
		if day < 91 {
			return 2
		}
		if day < 121 {
			return 3
		}
		if day < 152 {
			return 4
		}
		if day < 182 {
			return 5
		}
		if day < 213 {
			return 6
		}
		if day < 244 {
			return 7
		}
		if day < 274 {
			return 8
		}
		if day < 305 {
			return 9
		}
		if day < 335 {
			return 10
		}
		return 11
	}
	if day < 31 {
		return 0
	}
	if day < 59 {
		return 1
	}
	if day < 90 {
		return 2
	}
	if day < 120 {
		return 3
	}
	if day < 151 {
		return 4
	}
	if day < 181 {
		return 5
	}
	if day < 212 {
		return 6
	}
	if day < 243 {
		return 7
	}
	if day < 273 {
		return 8
	}
	if day < 304 {
		return 9
	}
	if day < 334 {
		return 10
	}
	return 11
}

func DateFromTime(t float64) float64 {
	day := DayWithinYear(t)
	month := MonthFromTime(t)

	var inLeapYear float64 = 0
	if InLeapYear(t) {
		inLeapYear = 1
	}
	switch month {
	case 0:
		return day + 1
	case 1:
		return day - 30
	case 2:
		return day - 58 - inLeapYear
	case 3:
		return day - 89 - inLeapYear
	case 4:
		return day - 119 - inLeapYear
	case 5:
		return day - 150 - inLeapYear
	case 6:
		return day - 180 - inLeapYear
	case 7:
		return day - 211 - inLeapYear
	case 8:
		return day - 242 - inLeapYear
	case 9:
		return day - 272 - inLeapYear
	case 10:
		return day - 303 - inLeapYear
	case 11:
		return day - 333 - inLeapYear
	}
	return 0
}

func WeekDay(t float64) float64 {
	return math.Mod(Day(t)+4, 7)
}

func HourFromTime(t float64) float64 {
	return math.Mod(math.Floor(t/3600000), 24)
}

func MinFromTime(t float64) float64 {
	return math.Mod(math.Floor(t/MS_PER_MIN), 60)
}

func SecFromTime(t float64) float64 {
	return math.Mod(math.Floor(t/1000), 60)
}

func msFromTime(t float64) float64 {
	return math.Mod(t, 1000)
}

func GetNamedTimeZoneOffsetNanoseconds(tz string, t float64) int {
	// TODO
	return 0
}

func SystemTimeZoneIdentifier() string {
	// TODO
	return "UTC"
}

// 21.4.1.26
func UTC(t float64) float64 {
	if math.IsInf(t, 0) {
		return math.NaN()
	}
	offsetNs := 0
	offsetMs := math.Trunc(float64(offsetNs) / 1e6)
	return t - offsetMs
}

// 21.4.1.27
func MakeTime(hour, min, sec, ms float64) float64 {
	return hour*60*60*1000 + min*60*1000 + sec*1000 + ms
}

// 21.4.1.28
func MakeDay(year, month, date float64) float64 {
	return year*12 + month*30 + date
}

// 21.4.1.29
func MakeDate(day, time float64) float64 {
	return day*MS_PER_DAY + time
}

// 21.4.1.30
func MakeFullYear(year float64) float64 {
	if math.IsNaN(year) {
		return math.NaN()
	}
	// TODO
	truncated := year
	if truncated >= 0 && truncated <= 99 {
		return 1900 + truncated
	}
	return truncated
}

func TimeClip(time float64) float64 {
	if math.IsInf(time, 0) {
		return math.NaN()
	}
	return math.Round(time)
}

func NewDateConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if newTarget == nil {
			now := time.Now().UTC()
			return NewStringValue(ToDateString(float64(now.UnixNano())))
		}
		numberOfArgs := len(args)
		var dv float64
		if numberOfArgs == 0 {
			dv = float64(time.Now().UnixNano())
		} else if numberOfArgs == 1 {
			value := args[0]
			var tv float64

			if o, ok := ValueGetObject(value); ok {
				if date, ok := o.(*DateObject); ok {
					tv = date.Data
				} else {

				}
			} else {
				v := ToPrimitive(agent, value, PreferredTypeNumber)
				if sv, ok := ValueGet[*StringValue](v); ok {
					tv = DateTimeStringFormat(sv.Data)
				} else {
					tv = ToNumber(agent, v).Data
				}
			}
			dv = TimeClip(tv)
		} else {
			year := ToNumber(agent, args[0]).Data
			month := ToNumber(agent, args[1]).Data
			date := ToNumber(agent, args[2]).Data
			hour := 0.0
			minute := 0.0
			sec := 0.0
			ms := 0.0
			if numberOfArgs >= 3 {
				hour = ToNumber(agent, args[3]).Data
			}
			if numberOfArgs >= 4 {
				minute = ToNumber(agent, args[4]).Data
			}
			if numberOfArgs >= 5 {
				sec = ToNumber(agent, args[5]).Data
			}
			if numberOfArgs >= 6 {
				ms = ToNumber(agent, args[6]).Data
			}
			year = MakeFullYear(year)
			day := MakeDay(year, month, date)
			_time := MakeTime(hour, minute, sec, ms)
			finalDate := MakeDate(day, _time)
			dv = TimeClip(UTC(finalDate))
		}
		o := OrdinaryCreateFromConstructor(agent, newTarget, "%Date.prototype", nil)
		d := &DateObject{
			Object: o,
			Data:   dv,
		}
		return NewValueFromObject(d)
	}
	object := CreateBuiltinFunction(agent, behavior, 7, "Date", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	var utc BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		numberOfArgs := len(args)
		if numberOfArgs < 1 {
			return NewNumberValue(math.NaN())
		}
		year := ToNumber(agent, args[0]).Data
		month := 0.0
		date := 1.0
		hour := 0.0
		minute := 0.0
		sec := 0.0
		ms := 0.0
		if numberOfArgs >= 2 {
			month = ToNumber(agent, args[1]).Data
		}
		if numberOfArgs >= 3 {
			date = ToNumber(agent, args[2]).Data
		}
		if numberOfArgs >= 4 {
			hour = ToNumber(agent, args[3]).Data
		}
		if numberOfArgs >= 5 {
			minute = ToNumber(agent, args[4]).Data
		}
		if numberOfArgs >= 6 {
			sec = ToNumber(agent, args[5]).Data
		}
		if numberOfArgs >= 7 {
			ms = ToNumber(agent, args[6]).Data
		}
		year = MakeFullYear(year)
		day := MakeDay(year, month, date)
		_time := MakeTime(hour, minute, sec, ms)
		finalDate := MakeDate(day, _time)
		dv := TimeClip(UTC(finalDate))
		return NewNumberValue(dv)
	}
	var now BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		return NewNumberValue(float64(time.Now().UnixNano()))
	}

	DefineBuiltinFunction(object, "UTC", utc, 7, realm)
	DefineBuiltinFunction(object, "now", now, 0, realm)

	DefineBuiltinProperty(object, "prototype", &PropertyDescriptor{
		Value:        NewValueFromObject(realm.Intrinsics.DatePrototype),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinProperty(realm.Intrinsics.DatePrototype, "constructor", NewValueFromObject(object))

	return object
}

// 21.4.1.25
func LocalTime(tv float64) float64 {
	// TODO
	return tv
}

func DateTimeStringFormat(s string) float64 {
	// TODO
	return 0
}

func DateString(t float64) string {
	// TODO
	return ""
}
func TimeString(t float64) string {
	// TODO
	return ""
}
func TimeZoneString(tv float64) string {
	// TODO
	return ""
}

// 21.4.4.41.4
func ToDateString(tv float64) string {
	if math.IsNaN(tv) {
		return "Invalid Date"
	}
	t := LocalTime(tv)

	s, err := time.Parse(time.RFC3339, time.Unix(int64(t), 0).Format(time.RFC3339))
	if err != nil {
		return "Invalid Date"
	}
	// Use go standard formatter
	return s.Format(time.UnixDate)
}
