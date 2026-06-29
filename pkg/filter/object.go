package filter

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PageSize struct {
	Size   int    `json:"size,omitempty"`
	Number int    `json:"number,omitempty"`
	Next   bool   `json:"with_id,omitempty"`
	LastId string `json:"last_id,omitempty"`
}

type FilterOptions struct {
	qs            string
	Filter        map[string]Filter `json:"filter,omitempty"`
	RestrictField map[string]bool   `json:"restrict_filter,omitempty"`
	AliasField    map[string]string `json:"alias_field,omitempty"`
	Page          PageSize          `json:"page,omitempty"`
	GroupBy       []string          `json:"group_by,omitempty"`
	Sort          []SortFilter      `json:"sort,omitempty"`
	Search        string            `json:"search,omitempty"`
	Fields        []FieldProjection `json:"fields,omitempty"`
	pagination    bool              `json:"-"`
	group         bool              `json:"-"`
	timezone      string            `json:"-"`
}

type FieldProjection struct {
	FieldName string `json:"field_name"`
	IsValue   bool   `json:"is_value"`
	Value     string `json:"value"`
}

type SortFilter struct {
	Field string `json:"field"`
	Index int    `json:"index"`
}

func (f FilterOptions) Skip() int64 {
	return int64((f.Page.Number - 1) * f.Page.Size)
}

func (f FilterOptions) Limit() int64 {
	return int64(f.Page.Size)
}

func (f FilterOptions) LastId() bool {
	if f.Page.Next {
		return true
	}
	return false
}

func (f FilterOptions) Pagination() bool {
	return f.pagination
}

func (f *FilterOptions) SetPagination() {
	f.pagination = true
}

func (f *FilterOptions) SetTimezone(zone string) {
	f.timezone = zone
}

func (f FilterOptions) Timezone() string {
	return f.timezone
}

func (f *FilterOptions) CheckAlias(field string) string {
	if v, ok := f.AliasField[field]; ok {
		return v
	}
	return field
}

func (f *FilterOptions) SetFilter(field, operator string, value interface{}) {
	var filter Filter
	if f.Filter == nil {
		f.Filter = map[string]Filter{}
	}
	filter.OrOperator = false
	filter.Value = append(filter.Value, OperatorFilter{Field: field, Operator: operator, Value: value})
	f.Filter[field] = filter
}

type Lookup struct {
	Data LookupData `bson:"$lookup" json:"$lookup"`
}

type LookupData struct {
	From         string `bson:"from" json:"from"`
	LocalField   string `bson:"localField" json:"localField"`
	ForeignField string `bson:"foreignField" json:"foreignField"`
	As           string `bson:"as" json:"as"`
}

type LookupPipeline struct {
	From string            `bson:"from" json:"from"`
	Let  map[string]string `bson:"let" json:"let"`
	Pipe []interface{}     `bson:"pipeline" json:"pipeline"`
	As   string            `bson:"as" json:"as"`
}

type UnWindData struct {
	Path         string `bson:"path" json:"path"`
	Preserve     bool   `bson:"preserveNullAndEmptyArrays" json:"preserveNullAndEmptyArrays"`
	IncludeIndex string `bson:"includeArrayIndex,omitempty" json:"includeArrayIndex,omitempty"`
}

type UnWindSpecial struct {
	Data UnWindData `bson:"$unwind" json:"$unwind"`
}

type Filter struct {
	OrOperator bool
	Value      []OperatorFilter
}

type OperatorFilter struct {
	Field     string
	Operator  string
	Value     interface{}
	TypeArray bool
	timezone  string
}

func convertValue(value string) any {
	if numberRE.MatchString(value) {
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			return number
		}
	}
	if objectIdRE.MatchString(value) {
		if object_id, err := primitive.ObjectIDFromHex(value); err == nil {
			return object_id
		}
	}
	if uuidRE.MatchString(value) {
		if parsedUUID, err := uuid.Parse(value); err == nil {
			return parsedUUID
		}
	}
	if datetimeRE.MatchString(value) {
		if date, err := time.Parse(types.LayoutISODateTime, value); err == nil {
			return date
		}
	}
	if dateRE.MatchString(value) {
		if date, err := time.Parse(types.LayoutISO, value); err == nil {
			return date
		}
	}
	if timeRE.MatchString(value) {
		if date, err := time.Parse(types.LayoutTime, value); err == nil {
			return date
		}
	}
	if t, err := time.Parse(types.LayoutRFC3339, value); err == nil {
		return t.UTC()
	}
	if b, err := strconv.Atoi(value); err == nil {
		return b
	}
	if b, err := strconv.ParseInt(value, 10, 64); err == nil {
		return b
	}
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return value
}

func attachOffset(datetime, tz string) string {
	if datetime == "0001-01-01 00:00:00" {
		return datetime
	}
	if tz == "" {
		return datetime
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return datetime
	}

	_, offset := time.Now().In(loc).Zone()

	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}

	h := offset / 3600
	m := (offset % 3600) / 60

	return strings.Replace(datetime, " ", "T", 1) + fmt.Sprintf("%s%02d:%02d", sign, h, m)
}

func (o OperatorFilter) ConvertValue(timezone string) interface{} {
	if o.TypeArray {
		var value []interface{}
		for _, v := range o.Value.([]string) {
			if datetimeRE.MatchString(v) {
				log.Printf("[debug][filter value date time]: %v\n", v)
				v = attachOffset(v, timezone)
			}
			value = append(value, convertValue(v))
		}
		return value
	}
	if s, ok := o.Value.(string); ok {
		if datetimeRE.MatchString(s) {
			s = attachOffset(s, timezone)
		}
		return convertValue(s)
	}
	return o.Value
}

func (o *OperatorFilter) SetTimezone(zone string) {
	o.timezone = zone
}

func CheckObjectIdNil(id primitive.ObjectID) bool {
	if id == primitive.NilObjectID {
		return true
	}
	return false
}

func ArrToStr(ids []primitive.ObjectID) string {
	var str []string
	for _, v := range ids {
		str = append(str, v.Hex())
	}
	return strings.Join(str, ",")
}

func ArrStr(ids []types.ObjectID) string {
	var str []string
	for _, v := range ids {
		str = append(str, v.String())
	}
	return strings.Join(str, ",")
}
