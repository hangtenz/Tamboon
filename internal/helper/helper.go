package helper

import (
	"reflect"
	"strconv"
	"time"
)

func MustParseInt(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}

func MustParseInt64(s string) int64 {
	return int64(MustParseInt(s))
}

func MustParseMonth(s string) time.Month {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return time.Month(result)
}

func Clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}
