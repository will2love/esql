package esql

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xwb1989/sqlparser"
)

// ProcessFunc ...
// esql use ProcessFunc to process query key value macro
type ProcessFunc func(string) (string, error)

// FilterFunc ...
// esql use FilterFunc to determine whether a target is to be processed by ProcessFunc
// only those FilterFunc(colName) == true will be processed
type FilterFunc func(string) bool

// ESql ...
// ESql is used to hold necessary information that required in parsing
type ESql struct {
	filterKey    FilterFunc  // select the column we want to process key macro
	filterValue  FilterFunc  // select the column we want to process value macro
	processKey   ProcessFunc // if selected by filterKey, change the query name
	processValue ProcessFunc // if selected by filterValue, change the query value
	pageSize     int
	bucketNumber int
}

// SetDefault ...
// all members goes to default
// should not be called if there is potential race condition
func (e *ESql) SetDefault() {
	e.pageSize = DefaultPageSize
	e.bucketNumber = DefaultBucketNumber
	e.filterKey = nil
	e.filterValue = nil
	e.processKey = nil
	e.processValue = nil
}

// NewESql ... return a new default ESql
func NewESql() *ESql {
	return &ESql{
		pageSize:     DefaultPageSize,
		bucketNumber: DefaultBucketNumber,
		processKey:   nil,
		processValue: nil,
		filterKey:    nil,
		filterValue:  nil,
	}
}

// ProcessQueryKey ... set up user specified column name processing policy
// should not be called if there is potential race condition
func (e *ESql) ProcessQueryKey(filterArg FilterFunc, replaceArg ProcessFunc) {
	e.filterKey = filterArg
	e.processKey = replaceArg
}

// ProcessQueryValue ... set up user specified column value processing policy
// should not be called if there is potential race condition
func (e *ESql) ProcessQueryValue(filterArg FilterFunc, processArg ProcessFunc) {
	e.filterValue = filterArg
	e.processValue = processArg
}

// SetPageSize ... set the number of documents returned in a non-aggregation query
// should not be called if there is potential race condition
func (e *ESql) SetPageSize(pageSizeArg int) {
	e.pageSize = pageSizeArg
}

// SetBucketNum ... set the number of bucket returned in an aggregation query
// should not be called if there is potential race condition
func (e *ESql) SetBucketNum(bucketNumArg int) {
	e.bucketNumber = bucketNumArg
}

// ConvertPretty ...
// Transform sql to elasticsearch dsl, and prettify the output json
//
// usage:
//  - dsl, sortField, err := e.ConvertPretty(sql, pageParam1, pageParam2, ...)
//
// arguments:
//  - sql: the sql query needs conversion in string format
//  - pagination: variadic arguments that indicates es search_after for pagination
//
// return values:
//  - dsl: the elasticsearch dsl json style string
//  - sortField: string array that contains all column names used for sorting. useful for pagination.
//  - err: contains err information
func (e *ESql) ConvertPretty(sql string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	dsl, sortField, err = e.Convert(sql, pagination...)
	if err != nil {
		return "", nil, err
	}

	var prettifiedDSLBytes bytes.Buffer
	err = json.Indent(&prettifiedDSLBytes, []byte(dsl), "", "  ")
	if err != nil {
		return "", nil, err
	}
	return string(prettifiedDSLBytes.Bytes()), sortField, err
}

// Convert ...
// Transform sql to elasticsearch dsl string
//
// usage:
//  - dsl, sortField, err := e.Convert(sql, pageParam1, pageParam2, ...)
//
// arguments:
//  - sql: the sql query needs conversion in string format
//  - pagination: variadic arguments that indicates es search_after
//
// return values:
//	- dsl: the elasticsearch dsl json style string
//	- sortField: string array that contains all column names used for sorting. useful for pagination.
//  - err: contains err information
func (e *ESql) Convert(sql string, pagination ...interface{}) (dsl string, sortField []string, err error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", nil, err
	}

	//sql valid, start to handle
	switch stmt.(type) {
	case *sqlparser.Select:
		dsl, sortField, err = e.convertSelect(*(stmt.(*sqlparser.Select)), "", pagination...)
	default:
		err = fmt.Errorf(`esql: Queries other than select not supported`)
	}

	if err != nil {
		return "", nil, err
	}
	return dsl, sortField, nil
}
