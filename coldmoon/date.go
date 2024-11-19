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
	return object
}

// 21.4.1.3
func Day(t float64) float64 {
	// FIXME: use time package
	msPerDay := 86400000
	return t / float64(msPerDay)
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
	return day*86400000 + time
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

// 21.4.4.41.4
func ToDateString(tv float64) string {
	if math.IsNaN(tv) {
		return "Invalid Date"
	}
	_ = LocalTime(tv)
	// TODO
	return ""
}
