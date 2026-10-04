package main

import (
	"math"
	"reflect"
	"strconv"
)

func firstNonFiniteNumber(value any) string {
	return firstNonFiniteValue(reflect.ValueOf(value), "result")
}

func firstNonFiniteValue(value reflect.Value, path string) string {
	if !value.IsValid() {
		return ""
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return ""
		}
		return firstNonFiniteValue(value.Elem(), path)
	case reflect.Float32, reflect.Float64:
		number := value.Float()
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return path
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			if found := firstNonFiniteValue(value.Field(i), path+"."+field.Name); found != "" {
				return found
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if found := firstNonFiniteValue(value.Index(i), path+"["+strconv.Itoa(i)+"]"); found != "" {
				return found
			}
		}
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			if found := firstNonFiniteValue(iter.Value(), path+"["+iter.Key().String()+"]"); found != "" {
				return found
			}
		}
	}
	return ""
}
