package utils

import (
	"reflect"
	"strconv"
	"time"
)

func StructToMapStrStr(data interface{}) map[string]string {
	result := make(map[string]string) //	结果容器

	type_ := reflect.TypeOf(data)   //  类型 集合
	value_ := reflect.ValueOf(data) //	值 集合

	for i := 0; i < value_.NumField(); i++ {
		field := type_.Field(i)
		fieldValue := value_.Field(i)

		value := fieldValue.Interface()

		switch value.(type) {
		case int:
			result[field.Name] = strconv.Itoa(value.(int))
		case string:
			result[field.Name] = fieldValue.Interface().(string)
		case uint32:
			result[field.Name] = strconv.Itoa(int(value.(uint32)))
		case time.Time:
			result[field.Name] = fieldValue.Interface().(time.Time).Format(time.RFC3339)
		}
	}
	return result
}
