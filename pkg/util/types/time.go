package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func ISODayOfWeekZero(t time.Time) int {
	w := int(t.Weekday())
	if w == 0 {
		return 6 // Sunday=6
	}
	return w - 1 // Monday=0
}

type DateISO struct {
	Year   int
	Month  time.Month
	Day    int
	Hour   int
	Minute int
	Second int
}

func ToISODate(t time.Time) DateISO {
	return DateISO{
		Year:   t.Year(),
		Month:  t.Month(),
		Day:    t.Day(),
		Hour:   t.Hour(),
		Minute: t.Minute(),
		Second: t.Second(),
	}
}

func (d DateISO) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, 0, time.UTC)
}

// UnmarshalJSON implements json.Unmarshaler inferface.
func (d *DateISO) UnmarshalJSON(b []byte) error {
	t, err := time.Parse(iso8601, strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}
	d.Year, d.Month, d.Day = t.Date()
	d.Hour = t.Hour()
	d.Minute = t.Minute()
	d.Second = t.Second()
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (d DateISO) MarshalJSON() ([]byte, error) {
	var s string
	if d.IsZero() {
		return []byte(`""`), nil
	}
	s = fmt.Sprintf(`"%04d-%02d-%02d %02d:%02d:%02d"`, d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second)
	return []byte(s), nil
}

// UnmarshalBSON implements bson.Unmarshaler inferface.
func (d *DateISO) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	if btype != bsontype.DateTime {
		return fmt.Errorf("cannot unmarshal non-isodate bson value to ShortDate")
	}
	vr := bsonrw.NewBSONValueReader(btype, data)
	dec, err := bson.NewDecoder(vr)
	if err != nil {
		return err
	}
	var mytime time.Time
	err = dec.Decode(&mytime)
	if err != nil {
		return err
	}
	d.Year, d.Month, d.Day = mytime.Date()
	d.Hour = mytime.Hour()
	d.Minute = mytime.Minute()
	d.Second = mytime.Second()
	return nil
}

func (d DateISO) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.IsZero() {
		return bsontype.Null, nil, nil
	}
	return bson.MarshalValue(time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, 0, time.UTC))
}

func (d DateISO) String() string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second)
}

// to unix()
func (d DateISO) Unix() int64 {
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, 0, time.UTC).Unix()
}

func (d DateISO) IsZero() bool {
	if d.Year == 0 && int(d.Month) == 0 && d.Day == 0 && d.Hour == 0 && d.Minute == 0 && d.Second == 0 {
		return true
	}
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, 0, time.UTC).IsZero()
}

type DateTime struct {
	Year   int
	Month  time.Month
	Day    int
	Hour   int
	Minute int
}

func ToDateTime(t time.Time) DateTime {
	return DateTime{
		Year:   t.Year(),
		Month:  t.Month(),
		Day:    t.Day(),
		Hour:   t.Hour(),
		Minute: t.Minute(),
	}
}

func (d DateTime) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, 0, 0, time.UTC)
}

// UnmarshalJSON implements json.Unmarshaler inferface.
func (d *DateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = DateTime{}
		return nil
	}
	s := strings.Trim(string(b), `"`)
	if s == "" {
		*d = DateTime{}
		return nil
	}
	t, err := time.Parse(rfc822, s)
	if err != nil {
		return err
	}
	d.Year, d.Month, d.Day = t.Date()
	d.Hour = t.Hour()
	d.Minute = t.Minute()
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (d DateTime) MarshalJSON() ([]byte, error) {
	var s string
	if d.Year == 0 && int(d.Month) == 0 && d.Day == 0 && d.Hour == 0 && d.Minute == 0 {
		return []byte(`""`), nil
	}
	s = fmt.Sprintf(`"%04d-%02d-%02d %02d:%02d"`, d.Year, d.Month, d.Day, d.Hour, d.Minute)
	return []byte(s), nil
}

// UnmarshalBSON implements bson.Unmarshaler inferface.
func (d *DateTime) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	if btype != bsontype.DateTime {
		return fmt.Errorf("cannot unmarshal non-datetime bson value to ShortDate")
	}
	vr := bsonrw.NewBSONValueReader(btype, data)
	dec, err := bson.NewDecoder(vr)
	if err != nil {
		return err
	}
	var mytime time.Time
	err = dec.Decode(&mytime)
	if err != nil {
		return err
	}
	d.Year, d.Month, d.Day = mytime.Date()
	d.Hour = mytime.Hour()
	d.Minute = mytime.Minute()
	return nil
}

func (d DateTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.IsZero() {
		return bsontype.Null, nil, nil
	}
	return bson.MarshalValue(time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, 0, 0, time.UTC))
}

func (d DateTime) String() string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d", d.Year, d.Month, d.Day, d.Hour, d.Minute)
}

func (d DateTime) Unix() int64 {
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, 00, 0, time.UTC).Unix()
}

func (d DateTime) IsZero() bool {
	if d.Year == 0 && int(d.Month) == 0 && d.Day == 0 && d.Hour == 0 && d.Minute == 0 {
		return true
	}
	return time.Date(d.Year, d.Month, d.Day, d.Hour, d.Minute, 0, 0, time.UTC).IsZero()
}

type ShortDate struct {
	Year  int
	Month time.Month
	Day   int
}

func ToShortDate(t time.Time) ShortDate {
	return ShortDate{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}
}

func StringToShortDate(b string) (ShortDate, error) {
	t, err := time.Parse(rfc3339, b)
	if err != nil {
		return ShortDate{}, err
	}
	return ShortDate{
		Year:  t.Year(),
		Month: t.Month(),
		Day:   t.Day(),
	}, nil
}

func (d *ShortDate) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func (d *ShortDate) UnmarshalText(b []byte) error {
	var t time.Time
	var err error

	if strings.Trim(string(b), `"`) != "" && len(b) > 0 {
		t, err = time.Parse(rfc3339, strings.Trim(string(b), `"`))
		if err != nil {
			return fmt.Errorf("cannot unmarshal non-datetime bson value to ShortDate %v, %s", err, b)
		}
	}
	d.Year, d.Month, d.Day = t.Date()
	return nil
}

func (d *ShortDate) UnmarshalJSON(b []byte) error {
	var t time.Time
	var err error
	if strings.Trim(string(b), `"`) != "" && strings.Trim(string(b), `"`) != "null" {
		t, err = time.Parse(rfc3339, strings.Trim(string(b), `"`))
		if err != nil {
			return fmt.Errorf("cannot unmarshal non-datetime bson value to ShortDate %v, %s", err, b)
		}
	}
	d.Year, d.Month, d.Day = t.Date()
	return nil
}

func (d ShortDate) MarshalJSON() ([]byte, error) {
	var s string
	if d.IsZero() {
		return []byte(`""`), nil
	}
	s = fmt.Sprintf(`"%04d-%02d-%02d"`, d.Year, d.Month, d.Day)
	return []byte(s), nil
}

func (d *ShortDate) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	if btype != bsontype.DateTime {
		return fmt.Errorf("cannot unmarshal non-datetime bson value to ShortDate")
	}
	vr := bsonrw.NewBSONValueReader(btype, data)
	dec, err := bson.NewDecoder(vr)
	if err != nil {
		return err
	}
	var mytime time.Time
	err = dec.Decode(&mytime)
	if err != nil {
		return err
	}
	if !mytime.IsZero() {
		d.Year, d.Month, d.Day = mytime.Date()
	}
	return nil
}

func (d ShortDate) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.IsZero() {
		return bsontype.Null, nil, nil
	}
	return bson.MarshalValue(time.Date(d.Year, d.Month, d.Day, 00, 0, 0, 0, time.UTC))
}

func (d ShortDate) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d ShortDate) Unix() int64 {
	return time.Date(d.Year, d.Month, d.Day, 00, 00, 00, 0, time.UTC).Unix()
}

func (d ShortDate) IsZero() bool {
	if d.Year == 0 && int(d.Month) == 0 && d.Day == 0 {
		return true
	}
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).IsZero()
}

type ShortTime struct {
	Hour int
	Min  int
	Zero bool
}

func ToShortTime(t time.Time) ShortTime {
	return ShortTime{
		Hour: t.Hour(),
		Min:  t.Minute(),
	}
}

func (d *ShortTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = ShortTime{Zero: true}
		return nil
	}
	s := strings.Trim(string(b), `"`)
	if s == "" {
		*d = ShortTime{Zero: true}
		return nil
	}
	t, err := time.Parse(kitchen, s)
	if err != nil {
		return err
	}

	*d = ShortTime{
		Hour: t.Hour(),
		Min:  t.Minute(),
		Zero: false,
	}
	return nil
}

func (d ShortTime) MarshalJSON() ([]byte, error) {
	// log.Println(d)
	s := `""`
	if !d.IsZero() {
		s = fmt.Sprintf(`"%02d:%02d"`, d.Hour, d.Min)
	}
	return []byte(s), nil
}

func (d *ShortTime) UnmarshalBSONValue(btype bsontype.Type, data []byte) error {
	*d = ShortTime{Hour: 0, Min: 0, Zero: true}
	switch btype {
	case bsontype.Null:
		*d = ShortTime{Hour: 0, Min: 0, Zero: true}
		return nil

	case bsontype.String:
		s, _, ok := bsoncore.ReadString(data)
		if !ok || s == "" {
			*d = ShortTime{Hour: 0, Min: 0, Zero: true}
			return nil
		}
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid time string %q", s)
		}
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		*d = ShortTime{Hour: h, Min: m, Zero: false}
		return nil

	case bsontype.DateTime:
		ms, _, ok := bsoncore.ReadDateTime(data)
		if !ok {
			return fmt.Errorf("invalid datetime")
		}
		tm := time.UnixMilli(ms).UTC()
		*d = ShortTime{Hour: tm.Hour(), Min: tm.Minute(), Zero: false}
		if tm.IsZero() {
			*d = ShortTime{Hour: tm.Hour(), Min: tm.Minute(), Zero: true}
		}
		return nil

	default:
		return fmt.Errorf("cannot unmarshal bson type %v to ShortTime", btype)
	}
}

func (d ShortTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.Zero {
		return bsontype.Null, nil, nil
	}
	return bson.MarshalValue(time.Date(0, 1, 1, d.Hour, d.Min, 0, 0, time.UTC))
}

func (d ShortTime) IsZero() bool {
	return d.Zero
}

func (d ShortTime) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", d.Hour, d.Min)
}

func CombineToTime(d ShortDate, t ShortTime) time.Time {
	if t.Zero {
		return time.Time{}
	}
	return time.Date(d.Year, d.Month, d.Day, t.Hour, t.Min, 0, 0, time.UTC)
}

func Midpoint(a, b time.Time) time.Time {
	if b.Before(a) {
		a, b = b, a
	}
	return a.Add(b.Sub(a) / 2)
}
func (s ShortTime) ToMinutes() int {
	return s.Hour*60 + s.Min
}
