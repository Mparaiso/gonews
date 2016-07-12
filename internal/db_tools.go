// @Copyright (c) 2016 mparaiso <mparaiso@online.fr>  All rights reserved.

package gonews

import (
	"fmt"
	"reflect"
)

type RowsScanner interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(destination ...interface{}) error
}

// Scanner populates destination values
// or returns an error
type Scanner interface {
	Scan(destination ...interface{}) error
}

// MapRowsToSliceOfSlices maps db rows to a slice of slices
func MapRowsToSliceOfSlices(scanner RowsScanner, Slices *[][]interface{}) error {
	defer scanner.Close()
	for scanner.Next() {
		columns, err := scanner.Columns()
		if err != nil {
			return err
		}
		sliceOfResults := make([]interface{}, len(columns))
		for i := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// create a new interface{} than is not nil
			sliceOfResults[i] = new(interface{})
		}
		err = scanner.Scan(sliceOfResults...)
		if err != nil {
			return err
		}
		row := []interface{}{}
		for index := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// convert the sliceOfResults[index] value back to interface{}
			var v interface{} = *(sliceOfResults[index].(*interface{}))
			if u, ok := v.([]uint8); ok {
				// v is likely a string, so convert to string
				row = append(row, (interface{}(string(u))))
			} else {
				row = append(row, v)
			}
		}
		*Slices = append(*Slices, row)
	}
	return scanner.Err()
}

// MapRowsToSliceOfMaps maps db rows to maps
// the map keys are the column names (or the aliases if defined in the query)
func MapRowsToSliceOfMaps(scanner RowsScanner, Map *[]map[string]interface{}) error {
	defer scanner.Close()
	for scanner.Next() {
		columns, err := scanner.Columns()
		if err != nil {
			return err
		}
		sliceOfResults := make([]interface{}, len(columns))
		for i := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// create a new interface{} than is not nil
			sliceOfResults[i] = new(interface{})
		}
		err = scanner.Scan(sliceOfResults...)
		if err != nil {
			return err
		}
		row := map[string]interface{}{}
		for index, column := range columns {
			// @see https://github.com/jmoiron/sqlx/blob/398dd5876282499cdfd4cb8ea0f31a672abe9495/sqlx.go#L751
			// convert the sliceOfResults[index] value back to interface{}
			var v interface{} = *(sliceOfResults[index].(*interface{}))
			if u, ok := v.([]uint8); ok {
				// v is likely a string, so convert to string
				row[column] = (interface{}(string(u)))
			} else {
				row[column] = v
			}
		}
		*Map = append(*Map, row)
	}
	return scanner.Err()
}

// MapRowsToSliceOfStruct  maps db rows to structs
func MapRowsToSliceOfStruct(scanner RowsScanner, sliceOfStructs interface{}, ignoreMissingField bool) error {
	///return connection.db.Select(records, query, parameters...)
	recordsPointerValue := reflect.ValueOf(sliceOfStructs)
	if recordsPointerValue.Kind() != reflect.Ptr {
		return fmt.Errorf("Expect pointer, got %#v", sliceOfStructs)
	}
	recordsValue := recordsPointerValue.Elem()
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("The underlying type is not a slice,pointer to slice expected for %#v ", recordsValue)
	}
	defer scanner.Close()
	columns, err := scanner.Columns()
	if err != nil {
		return err
	}
	// get the underlying type of a slice
	// @see http://stackoverflow.com/questions/24366895/golang-reflect-slice-underlying-type
	for scanner.Next() {
		//
		var t reflect.Type
		if recordsValue.Type().Elem().Kind() == reflect.Ptr {
			// the sliceOfStructs type is like []*T
			t = recordsValue.Type().Elem().Elem()
		} else {
			// the sliceOfStructs type is like []T
			t = recordsValue.Type().Elem()
		}
		pointerOfElement := reflect.New(t)

		err = MapRowToStruct(columns, scanner, pointerOfElement.Interface(), ignoreMissingField)
		if err != nil {
			return err
		}
		recordsValue = reflect.Append(recordsValue, pointerOfElement)
	}
	recordsPointerValue.Elem().Set(recordsValue)
	return scanner.Err()
}

// MapRowToStruct  automatically maps a db row to a struct .
func MapRowToStruct(columns []string, scanner Scanner, Struct interface{}, ignoreMissingFields bool) error {
	structPointer := reflect.ValueOf(Struct)
	if structPointer.Kind() != reflect.Ptr {
		return fmt.Errorf("Pointer expected, got %#v", Struct)
	}
	structValue := reflect.Indirect(structPointer)
	zeroValue := reflect.Value{}
	arrayOfResults := []interface{}{}
	for _, column := range columns {
		field := structValue.FieldByName(column)
		if field == zeroValue {
			if ignoreMissingFields {
				pointer := reflect.New(reflect.TypeOf([]byte{}))
				pointer.Elem().Set(reflect.ValueOf([]byte{}))
				arrayOfResults = append(arrayOfResults, pointer.Interface())

			} else {
				return fmt.Errorf("No field found for column %s in struct %#v", column, Struct)

			}
		} else {
			if !field.CanSet() {
				return fmt.Errorf("Unexported field %s cannot be set in struct %#v", column, Struct)
			}
			arrayOfResults = append(arrayOfResults, field.Addr().Interface())
		}
	}
	err := scanner.Scan(arrayOfResults...)
	if err != nil {
		return err
	}
	return nil
}
