package web_test

import (
	"net/url"
	"reflect"
	"testing"

	"forzatelemetry/storage"
	"forzatelemetry/web"

	"github.com/uptrace/bun"
)

type testParseFiltersRun struct {
	reqParam url.Values
	valid    []web.Filter
	missing  []string
	result   []storage.Where
	err      *web.ErrorRenderer
}

var valid = []web.Filter{
	web.MakeFilter("bool", "boolColumn", "bool", []string{"eq", "neq"}, ""),
	web.MakeFilter("int32", "int32Column", "int32", []string{"eq", "neq", "gt", "ge", "lt", "le"}, ""),
	web.MakeFilter("string", "stringColumn", "string", []string{"eq", "neq"}, ""),
	web.MakeFilter("[]string", "[]stringColumn", "[]string", []string{"in"}, ""),
	web.MakeFilter("[]int", "[]intColumn", "[]int", []string{"in"}, ""),

	web.MakeFilter("unexistantType", "[]intColumn", "unexistantType", []string{"eq"}, ""),
	web.MakeFilter("unexistantOperator", "boolColumn", "bool", []string{"unexistantOperator"}, ""),
}

var missing = []string{}

func TestParseGetRacesFilters(t *testing.T) {

	runs := map[string]testParseFiltersRun{
		"missing": {
			nil,
			valid,
			[]string{"bool:eq:true"},
			[]storage.Where{{Column: bun.Ident("boolColumn"), Operator: bun.Safe("="), Value: true}},
			nil,
		},
		"invalid": {
			url.Values{"filter": []string{"invalid"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'invalid': invalid syntax", nil, nil),
		},
		"invalid:name": {
			url.Values{"filter": []string{"invalid:eq:true"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'invalid:eq:true': invalid name", nil, nil),
		},
		"invalid:operator": {
			url.Values{"filter": []string{"bool:le:true"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'bool:le:true': invalid operator", nil, nil),
		},
		"unexistant:type": {
			url.Values{"filter": []string{"unexistantType:eq:true"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(500, "invalid filter 'unexistantType:eq:true': internal error", nil, nil),
		},
		"unexistant:operator": {
			url.Values{"filter": []string{"unexistantOperator:unexistantOperator:true"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(500, "invalid filter 'unexistantOperator:unexistantOperator:true': internal error", nil, nil),
		},
		"bool:valid:eq:true": {
			url.Values{"filter": []string{"bool:eq:true"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("boolColumn"), Operator: bun.Safe("="), Value: true}},
			nil,
		},
		"bool:valid:neq:true": {
			url.Values{"filter": []string{"bool:neq:true"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("boolColumn"), Operator: bun.Safe("!="), Value: true}},
			nil,
		},
		"bool:valid:false": {
			url.Values{"filter": []string{"bool:eq:false"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("boolColumn"), Operator: bun.Safe("="), Value: false}},
			nil,
		},
		"bool:invalid:aaa": {
			url.Values{"filter": []string{"bool:eq:aaa"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'bool:eq:aaa': invalid value: invalid syntax", nil, nil),
		},
		"bool:invalid:2": {
			url.Values{"filter": []string{"bool:eq:2"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'bool:eq:2': invalid value: invalid syntax", nil, nil),
		},
		"int32:valid:eq": {
			url.Values{"filter": []string{"int32:eq:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("int32Column"), Operator: bun.Safe("="), Value: int64(1)}},
			nil,
		},
		"int32:valid:neq": {
			url.Values{"filter": []string{"int32:neq:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("int32Column"), Operator: bun.Safe("!="), Value: int64(1)}},
			nil,
		},
		"int32:valid:gt": {
			url.Values{"filter": []string{"int32:gt:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("int32Column"), Operator: bun.Safe(">"), Value: int64(1)}},
			nil,
		},
		"int32:valid:ge": {
			url.Values{"filter": []string{"int32:ge:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("int32Column"), Operator: bun.Safe(">="), Value: int64(1)}},
			nil,
		},
		"int32:valid:lt": {
			url.Values{"filter": []string{"int32:lt:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("int32Column"), Operator: bun.Safe("<"), Value: int64(1)}},
			nil,
		},
		"int32:valid:le": {
			url.Values{"filter": []string{"int32:le:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("int32Column"), Operator: bun.Safe("<="), Value: int64(1)}},
			nil,
		},
		"int32:invalid:aaa": {
			url.Values{"filter": []string{"int32:eq:aaa"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'int32:eq:aaa': invalid value: invalid syntax", nil, nil),
		},
		"int32:invalid:9999999999": {
			url.Values{"filter": []string{"int32:eq:9999999999"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter 'int32:eq:9999999999': invalid value: value out of range", nil, nil),
		},
		"string:valid": {
			url.Values{"filter": []string{"string:eq:aaa"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("stringColumn"), Operator: bun.Safe("="), Value: "aaa"}},
			nil,
		},
		"[]string:valid:one": {
			url.Values{"filter": []string{"[]string:in:aaa"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("[]stringColumn"), Operator: bun.Safe("IN"), Value: []string{"aaa"}}},
			nil,
		},
		"[]string:valid:multiple": {
			url.Values{"filter": []string{"[]string:in:aaa,bbb"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("[]stringColumn"), Operator: bun.Safe("IN"), Value: []string{"aaa", "bbb"}}},
			nil,
		},
		"[]string:valid:,,,": {
			url.Values{"filter": []string{"[]string:in:,,,"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("[]stringColumn"), Operator: bun.Safe("IN"), Value: []string{"", "", "", ""}}},
			nil,
		},
		"[]int:valid:1": {
			url.Values{"filter": []string{"[]int:in:1"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("[]intColumn"), Operator: bun.Safe("IN"), Value: []int64{1}}},
			nil,
		},
		"[]int:valid:multiple": {
			url.Values{"filter": []string{"[]int:in:1,10"}},
			valid,
			missing,
			[]storage.Where{{Column: bun.Ident("[]intColumn"), Operator: bun.Safe("IN"), Value: []int64{1, 10}}},
			nil,
		},
		"[]int:invalid:aaa": {
			url.Values{"filter": []string{"[]int:in:aaa"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter '[]int:in:aaa': invalid value 'aaa': invalid syntax", nil, nil),
		},
		"[]int:invalid:9999999999": {
			url.Values{"filter": []string{"[]int:in:9999999999"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter '[]int:in:9999999999': invalid value '9999999999': value out of range", nil, nil),
		},
		"[]int:invalid:multiple:1,aaa": {
			url.Values{"filter": []string{"[]int:in:1,aaa"}},
			valid,
			missing,
			nil,
			web.NewErrorRenderer(400, "invalid filter '[]int:in:1,aaa': invalid value 'aaa': invalid syntax", nil, nil),
		},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			result, err := web.ParseFilters(run.reqParam, run.valid, run.missing)

			if !compareErrRenderer(err, run.err) {
				t.Errorf("expected %+v got %+v", run.err, err)
			}

			if !reflect.DeepEqual(result, run.result) {
				t.Errorf("expected %+v got %+v", run.result, result)
			}
		})
	}

}

func compareErrRenderer(actual *web.ErrorRenderer, expected *web.ErrorRenderer) bool {
	if actual == nil && expected == nil {
		return true
	}

	return actual.Msg == expected.Msg && actual.Status == expected.Status
}
