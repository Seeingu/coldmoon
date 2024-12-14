package coldmoon

import (
	"math"
	"time"
)

type DateObject struct {
	*Object
	Data JSNumber
}

func NewDatePrototype(realm *Realm) ObjectType {
	object := NewObject(realm.Agent, realm.Intrinsics.ObjectPrototype, "DatePrototype")
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
		if !tv.IsInf() {
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
		tv := ToPrimitive(agent, (o).ToValue(), PreferredTypeNumber)
		if n, ok := ValueGet[*NumberValue](tv); ok && !n.IsFinite() {
			return NullValue
		}
		return ValueInvoke(agent, (o).ToValue(), NewStringPropertyKey("toISOString"), nil)
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
		if tv.IsNaN() {
			return NewStringValue("Invalid Date")
		}
		return NewStringValue(DateString(LocalTime(tv)))
	}
	var toTimeString BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := MustGetObject(this).(*DateObject)
		tv := dateObject.Data
		if tv.IsNaN() {
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
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(tv - LocalTime(tv)/MS_PER_MIN)
	}
	var getDate BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(DateFromTime(tv))
	}
	var getDay BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(WeekDay(tv))
	}
	var getFullYear BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(YearFromTime(LocalTime(tv)))
	}
	var getHours BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(HourFromTime(LocalTime(tv)))
	}
	var getMilliseconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(msFromTime(LocalTime(tv)))
	}
	var getMinutes BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(MinFromTime(LocalTime(tv)))
	}
	var getMonth BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(MonthFromTime(LocalTime(tv)))
	}
	var getSeconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(SecFromTime(LocalTime(tv)))
	}
	var getTime BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		return NewNumberValue(tv)
	}
	var getUTCDate BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(DateFromTime(tv))
	}
	var getUTCDay BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(WeekDay(tv))
	}
	var getUTCFullYear BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(YearFromTime(tv))
	}
	var getUTCHours BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(HourFromTime(tv))
	}
	var getUTCMilliseconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(msFromTime(tv))
	}
	var getUTCMinutes BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(MinFromTime(tv))
	}
	var getUTCMonth BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(MonthFromTime(tv))
	}
	var getUTCSeconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		if tv.IsNaN() {
			return NaNValue
		}
		return NewNumberValue(SecFromTime(tv))
	}
	var setDate BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		date := ToNumber(agent, args[0]).Data
		if tv.IsNaN() {
			return NaNValue
		}
		t := LocalTime(tv)
		newDate := MakeDate(MakeDay(YearFromTime(t), MonthFromTime(t), date), TimeWithinDay(t))
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setFullYear BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		year := ToNumber(agent, args[0]).Data
		month := JSNumber(0.0)
		date := JSNumber(1.0)
		if len(args) >= 2 {
			month = ToNumber(agent, args[1]).Data
		}
		if len(args) >= 3 {
			date = ToNumber(agent, args[2]).Data
		}
		year = MakeFullYear(year)
		day := MakeDay(year, month, date)
		newDate := MakeDate(day, TimeWithinDay(tv))
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setHours BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		hour := ToNumber(agent, args[0]).Data
		minute := JSNumber(0.0)
		sec := JSNumber(0.0)
		ms := JSNumber(0.0)
		if len(args) >= 2 {
			minute = ToNumber(agent, args[1]).Data
		}
		if len(args) >= 3 {
			sec = ToNumber(agent, args[2]).Data
		}
		if len(args) >= 4 {
			ms = ToNumber(agent, args[3]).Data
		}
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(hour, minute, sec, ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setMilliseconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		ms := ToNumber(agent, args[0]).Data
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(HourFromTime(tv), MinFromTime(tv), SecFromTime(tv), ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setMinutes BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		minute := ToNumber(agent, args[0]).Data
		sec := JSNumber(0.0)
		ms := JSNumber(0.0)
		if len(args) >= 2 {
			sec = ToNumber(agent, args[1]).Data
		}
		if len(args) >= 3 {
			ms = ToNumber(agent, args[2]).Data
		}
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(HourFromTime(tv), minute, sec, ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setMonth BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		month := ToNumber(agent, args[0]).Data
		date := JSNumber(1.0)
		if len(args) >= 2 {
			date = ToNumber(agent, args[1]).Data
		}
		day := MakeDay(YearFromTime(tv), month, date)
		newDate := MakeDate(day, TimeWithinDay(tv))
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setSeconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		sec := ToNumber(agent, args[0]).Data
		ms := JSNumber(0.0)
		if len(args) >= 2 {
			ms = ToNumber(agent, args[1]).Data
		}
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(HourFromTime(tv), MinFromTime(tv), sec, ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setTime BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		t := ToNumber(agent, args[0]).Data
		dateObject.Data = TimeClip(t)
		return NewNumberValue(dateObject.Data)
	}
	var setUTCDate BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		date := ToNumber(agent, args[0]).Data
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), date)
		newDate := MakeDate(day, TimeWithinDay(tv))
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setUTCHours BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		hour := ToNumber(agent, args[0]).Data
		minute := JSNumber(0.0)
		sec := JSNumber(0.0)
		ms := JSNumber(0.0)
		if len(args) >= 2 {
			minute = ToNumber(agent, args[1]).Data
		}
		if len(args) >= 3 {
			sec = ToNumber(agent, args[2]).Data
		}
		if len(args) >= 4 {
			ms = ToNumber(agent, args[3]).Data
		}
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(hour, minute, sec, ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setUTCMilliseconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		ms := ToNumber(agent, args[0]).Data
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(HourFromTime(tv), MinFromTime(tv), SecFromTime(tv), ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setUTCMinutes BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		minute := ToNumber(agent, args[0]).Data
		sec := JSNumber(0.0)
		ms := JSNumber(0.0)
		if len(args) >= 2 {
			sec = ToNumber(agent, args[1]).Data
		}
		if len(args) >= 3 {
			ms = ToNumber(agent, args[2]).Data
		}
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(HourFromTime(tv), minute, sec, ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setUTCMonth BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		month := ToNumber(agent, args[0]).Data
		date := JSNumber(1.0)
		if len(args) >= 2 {
			date = ToNumber(agent, args[1]).Data
		}
		day := MakeDay(YearFromTime(tv), month, date)
		newDate := MakeDate(day, TimeWithinDay(tv))
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
	}
	var setUTCSeconds BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		dateObject := RequireInternalSlot[*DateObject](this)
		tv := dateObject.Data
		sec := ToNumber(agent, args[0]).Data
		ms := JSNumber(0.0)
		if len(args) >= 2 {
			ms = ToNumber(agent, args[1]).Data
		}
		day := MakeDay(YearFromTime(tv), MonthFromTime(tv), DateFromTime(tv))
		newTime := MakeTime(HourFromTime(tv), MinFromTime(tv), sec, ms)
		newDate := MakeDate(day, newTime)
		u := TimeClip(UTC(newDate))
		dateObject.Data = u
		return NewNumberValue(u)
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
	DefineBuiltinFunction(object, "getDate", getDate, 0, realm)
	DefineBuiltinFunction(object, "getDay", getDay, 0, realm)
	DefineBuiltinFunction(object, "getFullYear", getFullYear, 0, realm)
	DefineBuiltinFunction(object, "getHours", getHours, 0, realm)
	DefineBuiltinFunction(object, "getMilliseconds", getMilliseconds, 0, realm)
	DefineBuiltinFunction(object, "getMinutes", getMinutes, 0, realm)
	DefineBuiltinFunction(object, "getMonth", getMonth, 0, realm)
	DefineBuiltinFunction(object, "getSeconds", getSeconds, 0, realm)
	DefineBuiltinFunction(object, "getTime", getTime, 0, realm)
	DefineBuiltinFunction(object, "getUTCDate", getUTCDate, 0, realm)
	DefineBuiltinFunction(object, "getUTCDay", getUTCDay, 0, realm)
	DefineBuiltinFunction(object, "getUTCFullYear", getUTCFullYear, 0, realm)
	DefineBuiltinFunction(object, "getUTCHours", getUTCHours, 0, realm)
	DefineBuiltinFunction(object, "getUTCMilliseconds", getUTCMilliseconds, 0, realm)
	DefineBuiltinFunction(object, "getUTCMinutes", getUTCMinutes, 0, realm)
	DefineBuiltinFunction(object, "getUTCMonth", getUTCMonth, 0, realm)
	DefineBuiltinFunction(object, "getUTCSeconds", getUTCSeconds, 0, realm)
	DefineBuiltinFunction(object, "setDate", setDate, 1, realm)
	DefineBuiltinFunction(object, "setFullYear", setFullYear, 3, realm)
	DefineBuiltinFunction(object, "setHours", setHours, 4, realm)
	DefineBuiltinFunction(object, "setMilliseconds", setMilliseconds, 1, realm)
	DefineBuiltinFunction(object, "setMinutes", setMinutes, 3, realm)
	DefineBuiltinFunction(object, "setMonth", setMonth, 2, realm)
	DefineBuiltinFunction(object, "setSeconds", setSeconds, 2, realm)
	DefineBuiltinFunction(object, "setTime", setTime, 1, realm)
	DefineBuiltinFunction(object, "setUTCDate", setUTCDate, 1, realm)
	DefineBuiltinFunction(object, "setUTCHours", setUTCHours, 4, realm)
	DefineBuiltinFunction(object, "setUTCMilliseconds", setUTCMilliseconds, 1, realm)
	DefineBuiltinFunction(object, "setUTCMinutes", setUTCMinutes, 3, realm)
	DefineBuiltinFunction(object, "setUTCMonth", setUTCMonth, 2, realm)
	DefineBuiltinFunction(object, "setUTCSeconds", setUTCSeconds, 2, realm)

	DefineBuiltinFunctionWithAttributes(object, "@@toPrimitive", toPrimitive, 1, realm, PropertyDescriptorAttributes{
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	return object
}

const (
	MS_PER_DAY = JSNumber(86400000)
	MS_PER_MIN = JSNumber(60000)
)

// 21.4.1.3
func Day(t JSNumber) JSNumber {
	// FIXME: use time package
	msPerDay := MS_PER_DAY
	return t / msPerDay
}

func TimeWithinDay(t JSNumber) JSNumber {
	return t.Mod(MS_PER_DAY)
}

func DaysInYear(y JSNumber) JSNumber {
	if y.Mod(4) != 0 {
		return 365
	}
	if y.Mod(100) != 0 {
		return 366
	}
	if y.Mod(400) != 0 {
		return 365
	}
	return 366
}

func DayFromYear(y JSNumber) JSNumber {
	return 365*(y-1970) + ((y - 1969) / 4).Floor() - ((y - 1901) / 100).Floor() + ((y - 1601) / 400).Floor()
}

func TimeFromYear(y JSNumber) JSNumber {
	return MS_PER_DAY * DayFromYear(y)
}

func YearFromTime(t JSNumber) JSNumber {
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

func DayWithinYear(t JSNumber) JSNumber {
	return Day(t) - DayFromYear(YearFromTime(t))
}

func InLeapYear(t JSNumber) bool {
	if DaysInYear(YearFromTime(t)) == 366 {
		return true
	}
	return false
}

func MonthFromTime(t JSNumber) JSNumber {
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

func DateFromTime(t JSNumber) JSNumber {
	day := DayWithinYear(t)
	month := MonthFromTime(t)

	var inLeapYear JSNumber = 0
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

func WeekDay(t JSNumber) JSNumber {
	return (Day(t) + 4).Mod(7)
}

func HourFromTime(t JSNumber) JSNumber {
	return (t / 3600000).Floor().Mod(24)
}

func MinFromTime(t JSNumber) JSNumber {
	return (t / MS_PER_MIN).Floor().Mod(60)
}

func SecFromTime(t JSNumber) JSNumber {
	return JSNumber(math.Mod(math.Floor(t.ToFloat()/1000), 60))
}

func msFromTime(t JSNumber) JSNumber {
	return t.Mod(1000)
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
func UTC(t JSNumber) JSNumber {
	if t.IsInf() {
		return JSNumberNaN
	}
	offsetNs := 0
	offsetMs := math.Trunc(float64(offsetNs) / 1e6)
	return t - JSNumber(offsetMs)
}

// 21.4.1.27
func MakeTime(hour, min, sec, ms JSNumber) JSNumber {
	return hour*60*60*1000 + min*60*1000 + sec*1000 + ms
}

// 21.4.1.28
func MakeDay(year, month, date JSNumber) JSNumber {
	return year*12 + month*30 + date
}

// 21.4.1.29
func MakeDate(day, time JSNumber) JSNumber {
	return day*MS_PER_DAY + time
}

// 21.4.1.30
func MakeFullYear(year JSNumber) JSNumber {
	if year.IsNaN() {
		return JSNumberNaN
	}
	// TODO
	truncated := year
	if truncated >= 0 && truncated <= 99 {
		return 1900 + truncated
	}
	return truncated
}

func TimeClip(time JSNumber) JSNumber {
	if time.IsInf() {
		return JSNumberNaN
	}
	return JSNumber(math.Round(time.ToFloat()))
}

func NewDateConstructor(realm *Realm) ObjectType {
	agent := realm.Agent
	var behavior BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		if newTarget == nil {
			now := time.Now().UTC()
			return NewStringValue(ToDateString(JSNumber(now.UnixNano())))
		}
		numberOfArgs := len(args)
		var dv JSNumber
		if numberOfArgs == 0 {
			dv = JSNumber(time.Now().UnixNano())
		} else if numberOfArgs == 1 {
			value := args[0]
			var tv JSNumber

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
			hour := JSNumber(0.0)
			minute := JSNumber(0.0)
			sec := JSNumber(0.0)
			ms := JSNumber(0.0)
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
		return (d).ToValue()
	}
	object := CreateBuiltinFunction(agent, behavior, 7, "Date", builtinFunctionArgs{
		realm:     realm,
		prototype: realm.Intrinsics.FunctionPrototype,
	})

	var utc BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		numberOfArgs := len(args)
		if numberOfArgs < 1 {
			return NaNValue
		}
		year := ToNumber(agent, args[0]).Data
		month := JSNumber(0.0)
		date := JSNumber(1.0)
		hour := JSNumber(0.0)
		minute := JSNumber(0.0)
		sec := JSNumber(0.0)
		ms := JSNumber(0.0)
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
		return NewNumberValue(JSNumber(time.Now().UnixNano()))
	}
	var parse BehaviorFn = func(this Value, args []Value, newTarget ObjectType) Value {
		s := args[0].String()
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return NaNValue
		}
		return NewNumberValue(JSNumber(t.UnixNano()) / 1e6)
	}

	DefineBuiltinFunction(object, "UTC", utc, 7, realm)
	DefineBuiltinFunction(object, "now", now, 0, realm)
	DefineBuiltinFunction(object, "parse", parse, 1, realm)

	DefineBuiltinPropertyP(object, "prototype", &PropertyDescriptor{
		Value:        realm.Intrinsics.DatePrototype.ToValue(),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
	DefineBuiltinPropertyV(realm.Intrinsics.DatePrototype, "constructor", object.ToValue())

	return object
}

// 21.4.1.25
func LocalTime(tv JSNumber) JSNumber {
	// TODO
	return tv
}

func DateTimeStringFormat(s string) JSNumber {
	// TODO
	return 0
}

func DateString(t JSNumber) string {
	// TODO
	return ""
}

func TimeString(t JSNumber) string {
	// TODO
	return ""
}

func TimeZoneString(tv JSNumber) string {
	// TODO
	return ""
}

// 21.4.4.41.4
func ToDateString(tv JSNumber) string {
	if tv.IsNaN() {
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
