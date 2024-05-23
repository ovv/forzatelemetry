package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"forzatelemetry/storage"
)

var ErrInternal = fmt.Errorf("internal error")

type Filter struct {
	colum string

	Name      string   `json:"name"`
	Type      string   `json:"valueType"`
	Operators []string `json:"operators"`
	Example   string   `json:"example"`
}

func MakeFilter(name string, colum string, type_ string, operators []string, example string) Filter {
	return Filter{
		Name:      name,
		colum:     colum,
		Type:      type_,
		Operators: operators,
		Example:   example,
	}
}

func (f Filter) toWhereClause(operator string, value string) (storage.Where, error) {
	if !f.validateOperator(operator) {
		return storage.Where{}, fmt.Errorf("invalid operator")
	}

	var v any
	var err error
	switch f.Type {
	case "bool":
		v, err = strconv.ParseBool(value)
		if err != nil {
			return storage.Where{}, fmt.Errorf("invalid value: %s", err.(*strconv.NumError).Err)
		}
	case "int32":
		v, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return storage.Where{}, fmt.Errorf("invalid value: %s", err.(*strconv.NumError).Err)
		}
	case "time":
		ms, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return storage.Where{}, fmt.Errorf("invalid value: %s", err.(*strconv.NumError).Err)
		}
		v = time.UnixMilli(ms)
	case "string":
		v = value
	case "[]string":
		v = strings.Split(value, ",")
	case "[]int":
		var result []int64
		for _, val := range strings.Split(value, ",") {
			i, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				return storage.Where{}, fmt.Errorf("invalid value '%s': %s", val, err.(*strconv.NumError).Err)
			}
			result = append(result, i)
		}
		v = result
	default:
		return storage.Where{}, ErrInternal
	}

	var op bun.Safe
	switch operator {
	case "eq":
		op = bun.Safe("=")
	case "neq":
		op = bun.Safe("!=")
	case "gt":
		op = bun.Safe(">")
	case "ge":
		op = bun.Safe(">=")
	case "lt":
		op = bun.Safe("<")
	case "le":
		op = bun.Safe("<=")
	case "in":
		op = bun.Safe("IN")
	default:
		return storage.Where{}, ErrInternal
	}

	return storage.Where{
		Column:   bun.Ident(f.colum),
		Operator: op,
		Value:    v,
	}, nil
}

func (f Filter) validateOperator(operator string) bool {
	for _, op := range f.Operators {
		if op == operator {
			return true
		}
	}
	return false
}

func ParseFilters(param url.Values, valid []Filter, missing []string) ([]storage.Where, *ErrorRenderer) {
	var filters []storage.Where

	rawFilters, ok := param["filter"]
	if !ok {
		rawFilters = missing
	}

	for _, rawFilter := range rawFilters {
		filter, err := parseFilter(rawFilter, valid)
		if err != nil {
			return nil, filterParseError(rawFilter, err, valid)
		}
		if filter.Value != nil {
			filters = append(filters, filter)
		}
	}
	return filters, nil
}

func parseFilter(raw string, valid []Filter) (storage.Where, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return storage.Where{}, fmt.Errorf("invalid syntax")
	}
	name, operator, value := parts[0], parts[1], parts[2]

	if value == "*" {
		return storage.Where{}, nil
	}

	for _, f := range valid {
		if f.Name == name {
			return f.toWhereClause(operator, value)
		}
	}
	return storage.Where{}, fmt.Errorf("invalid name")
}

func filterParseError(filter string, err error, valid []Filter) *ErrorRenderer {
	status := http.StatusBadRequest
	if errors.Is(err, ErrInternal) {
		status = http.StatusInternalServerError
	}

	return NewErrorRenderer(
		status,
		fmt.Sprintf("invalid filter '%s': %s", filter, err.Error()),
		err,
		map[string]any{
			"doc":     "filters must be a string with a specific syntax and values",
			"syntax":  "<name>:<operator>:<value>",
			"filters": valid,
		},
	)
}
