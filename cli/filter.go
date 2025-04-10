package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"strconv"
	"strings"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/tidwall/gjson"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	COND_POSITIVE uint8 = 0
	COND_NEGATIVE uint8 = 1
	COND_CONTAINS uint8 = 2
	COND_LESS     uint8 = 3
	COND_GREATER  uint8 = 4
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Filter is input filter
type Filter struct {
	Key   string
	Value any
	Cond  uint8
}

// Filters is a slice of filters
type Filters []Filter

// Highlights is a slice of highlights
type Highlights []string

// ////////////////////////////////////////////////////////////////////////////////// //

var conditions = map[rune]uint8{
	'!': COND_NEGATIVE,
	'~': COND_CONTAINS,
	'<': COND_LESS,
	'>': COND_GREATER,
}

// ////////////////////////////////////////////////////////////////////////////////// //

// parseFilters parses raw filters data
func parseFilters(filters []string) Filters {
	var result Filters

	for _, f := range filters {
		result = append(result, parseFilter(f))
	}

	return result
}

// parseFilter parses raw filter string
func parseFilter(f string) Filter {
	key, value, ok := strings.Cut(f, ":")

	if !ok || key == "" || value == "" {
		return Filter{Key: "msg", Value: f, Cond: COND_CONTAINS}
	}

	filter := Filter{Key: key, Cond: conditions[rune(value[0])]}

	if filter.Cond != COND_POSITIVE {
		value = value[1:]
	}

	if filter.Cond == COND_GREATER || filter.Cond == COND_LESS {
		fv, _ := strconv.ParseFloat(value, 10)
		filter.Value = fv
	} else {
		filter.Value = value
	}

	return filter
}

// ////////////////////////////////////////////////////////////////////////////////// //

// IsMatch checks if json record fields match filters
func (f Filters) IsMatch(fields map[string]gjson.Result) bool {
	if len(f) == 0 {
		return true
	}

	for _, ff := range f {
		jf, ok := fields[ff.Key]

		if !ok {
			return false
		}

		switch ff.Cond {
		case COND_POSITIVE:
			if ff.Value.(string) != jf.String() {
				return false
			}

		case COND_NEGATIVE:
			if ff.Value.(string) == jf.String() {
				return false
			}

		case COND_CONTAINS:
			if !strings.Contains(jf.String(), ff.Value.(string)) {
				return false
			}

		case COND_GREATER:
			if ff.Value.(float64) > jf.Float() {
				return false
			}

		case COND_LESS:
			if ff.Value.(float64) < jf.Float() {
				return false
			}
		}
	}

	return true
}

// Apply applies highlights to given message
func (h Highlights) Apply(msg string) (string, bool) {
	var found bool

	for _, hh := range h {
		if strings.Contains(msg, hh) {
			msg = strings.ReplaceAll(msg, hh, fmtc.Sprint("{#112}{_}"+hh+"{!}"))
			found = true
		}
	}

	return msg, found
}
