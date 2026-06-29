package filter

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
)

func parseValueParams(key, qs string) ([]OperatorFilter, error) {
	var op []OperatorFilter
	terms := operatorRE.FindAllStringSubmatch(qs, -1)
	values := valueRE.FindAllStringSubmatch(qs, -1)

	if len(terms) > 0 && len(terms) > len(values) {
		return op, fmt.Errorf("unable to parse value: an object hierarchy has been provided")
	}
	var keys []string
	if commaRE.MatchString(key) {
		keys = commaRE.Split(key, -1)
	} else {
		keys = []string{key}
	}

	if len(terms) != len(keys) && len(keys) > 1 {
		return op, fmt.Errorf("unable to parse value: an object hierarchy has been provided")
	}
	for i, term := range terms {
		var o OperatorFilter
		o.Field = keys[0]
		if len(keys) > 1 {
			o.Field = keys[i]
		}
		o.Operator = strings.ToLower(term[1])
		if commaRE.MatchString(values[i][1]) || o.Operator == "in" {
			o.Value = commaRE.Split(values[i][1], -1)
			o.TypeArray = true
			op = append(op, o)
			continue
		}
		o.Value = values[i][1]
		op = append(op, o)
	}

	return op, nil
}

func parseSearch(qs string) string {
	matches := searchRE.FindAllStringSubmatch(qs, -1)
	var srt string
	if len(matches) <= 0 {
		return srt
	}
	for _, v := range matches {
		v_nw := strings.ReplaceAll(v[1], `"`, "")
		v_nw = strings.ReplaceAll(v_nw, `'`, "")
		if len(v_nw) > 0 {
			if srt == "" {
				srt = v[1]
			} else {
				srt = fmt.Sprintf("%s %s", srt, v[1])
			}
		}
	}
	return srt
}

func parseGroupBy(qs string) []string {
	matches := groupByRE.FindAllStringSubmatch(qs, -1)
	var keys []string
	if len(matches) <= 0 {
		return keys
	}
	for _, v := range matches {
		if commaRE.MatchString(v[1]) {
			keys = commaRE.Split(v[1], -1)
		} else {
			keys = []string{v[1]}
		}
	}
	return keys
}

func parseFieldProjection(query string) []FieldProjection {
	var result []FieldProjection

	if m := fieldsRE.FindStringSubmatch(query); len(m) > 1 {
		fields := strings.Split(m[1], ",")

		for _, f := range fields {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}

			result = append(result, FieldProjection{
				FieldName: f,
				IsValue:   false,
			})
		}
	}

	matches := fieldsWithRE.FindAllStringSubmatch(query, -1)

	for _, m := range matches {
		field := m[1]
		value := m[2]

		result = append(result, FieldProjection{
			FieldName: field,
			IsValue:   true,
			Value:     value,
		})
	}

	return result
}

func CheckParams(qs string, o *FilterOptions) error {
	var err error

	if qs == "" {
		return nil
	}
	qs, err = url.QueryUnescape(qs)
	if err != nil {
		return err
	}
	terms := bracketRE.FindAllStringSubmatch(qs, -1)
	values := bracketValueRE.FindAllStringSubmatch(qs, -1)

	if len(terms) > 0 && len(terms) > len(values) {
		return fmt.Errorf("unable to parse 1: an object hierarchy has been provided")
	}
	for _, term := range terms {
		switch strings.ToLower(term[1]) {
		case "filter":
			if _, ok := o.RestrictField[term[2]]; ok {
				return fmt.Errorf("unable to parse value: field %v is restrict", term[2])
			}
			if commaRE.MatchString(term[2]) {
				for _, v := range commaRE.Split(term[2], -1) {
					if _, ok := o.RestrictField[v]; ok {
						return fmt.Errorf("unable to parse value: field %s is restrict", v)
					}
				}
			}
		}
	}
	return nil
}

func cleanQuery(raw string) string {
	matches := cleanRE.FindAllString(raw, -1)
	cleaned := strings.Join(matches, "&")
	return cleaned
}

func ParseBracketParams(qs string, o *FilterOptions) error {
	o.Page.Size = 100
	o.Page.Number = 1
	var err error

	if qs == "" {
		return nil
	}
	qs, err = url.QueryUnescape(qs)
	if err != nil {
		return err
	}
	o.Search = parseSearch(qs)
	o.GroupBy = parseGroupBy(qs)
	o.Fields = parseFieldProjection(qs)
	if len(o.GroupBy) > 3 {
		return fmt.Errorf("unable to parse 3: an object hierarchy has been provided")
	}
	if len(o.GroupBy) > 0 {
		qs = cleanQuery(qs)
	}
	terms := bracketRE.FindAllStringSubmatch(qs, -1)
	values := bracketValueRE.FindAllStringSubmatch(qs, -1)

	if len(terms) > 0 && len(terms) > len(values) {
		return fmt.Errorf("unable to parse 1: an object hierarchy has been provided")
	}
	for i, term := range terms {
		switch strings.ToLower(term[1]) {
		case "filter":
			if o.Filter == nil {
				o.Filter = map[string]Filter{}
			}
			if f, ok := o.Filter[term[2]]; ok {
				f_values, err := parseValueParams(term[2], values[i][1])
				if err != nil {
					return err
				}
				f.Value = append(f.Value, f_values...)
				o.Filter[term[2]] = f
				continue
			}
			var filter Filter
			filter.OrOperator = false
			if commaRE.MatchString(term[2]) {
				filter.OrOperator = true
			}
			filter.Value, err = parseValueParams(term[2], values[i][1])
			if err != nil {
				return err
			}

			o.Filter[term[2]] = filter
		case "page":
			if !o.Pagination() {
				return fmt.Errorf("Cannot use pagination in api")
			}
			if term[2] == "size" {
				v, err := strconv.ParseInt(values[i][1], 0, 64)
				if err != nil {
					return fmt.Errorf("unable to parse: page size do not number")
				}
				if int(v) > 100000 {
					o.Page.Size = 100000
				} else {
					o.Page.Size = int(v)
				}
			}

			if term[2] == "number" {
				v, err := strconv.ParseInt(values[i][1], 0, 64)
				if err != nil {
					return fmt.Errorf("unable to parse: page number do not number")
				}
				o.Page.Number = int(v)
			}

			if term[2] == "next" {
				var id string
				id, err = helper.DecryptID(values[i][1])
				if err != nil {
					return fmt.Errorf("unable to parse: page next do not id")
				}
				o.Page.LastId = id
				o.Page.Next = true
			}

		case "sort":
			if values[i][1] == "desc" {
				o.Sort = append(o.Sort, SortFilter{Field: term[2], Index: -1})
				continue
			}
			o.Sort = append(o.Sort, SortFilter{Field: term[2], Index: 1})
		}
	}

	return nil
}
