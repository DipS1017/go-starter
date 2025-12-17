package utils

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/webpoint-solutions-llc/dba/internal/errorhandler"
)

func ConvertIntToNumeric(x int32) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric

	// Convert the integer to a string before scanning
	stringValue := fmt.Sprintf("%d", x)

	if err := numeric.Scan(stringValue); err != nil {
		log.Println("Error scanning value into pgtype.Numeric:", err)
		return pgtype.Numeric{}, err
	}

	return numeric, nil
}

func ConvertStringToNumeric(s string) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	err := num.Scan(s)
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("failed to scan string to pgtype.Numeric: %w", err)
	}
	return num, nil
}

func Float64ToPgNumeric(f float64) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric

	// Convert float64 to string
	str := fmt.Sprintf("%f", f)

	// Use Scan to convert the string to pgtype.Numeric
	err := numeric.Scan(str)
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("failed to scan string into pgtype.Numeric: %w", err)
	}

	return numeric, nil
}

func ConvertStringToTime(s string) (pgtype.Time, error) {
	var time pgtype.Time

	// Use Scan to convert the string to pgtype.Time
	err := time.Scan(s)
	if err != nil {
		return pgtype.Time{}, fmt.Errorf("failed to scan string into pgtype.Time: %w", err)
	}

	return time, nil
}

func CopyValidPointers(src interface{}, dest interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	destVal := reflect.ValueOf(dest).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		field := srcVal.Field(i)
		fieldType := srcVal.Type().Field(i)
		destField := destVal.FieldByName(fieldType.Name)

		if field.Kind() == reflect.Ptr && !field.IsNil() && destField.IsValid() && destField.CanSet() {
			destField.Set(reflect.Indirect(field))
		}
	}
}

func BoolPointer(b bool) *bool {
	return &b
}

func NormalizeStringArray(arr []string) []string {
	for i := range arr {
		arr[i] = strings.ToUpper(strings.TrimSpace(arr[i]))
	}
	return arr
}

func NormalizeIntArray(arr []string) (res []int32, err error) {
	var arrInt []int32
	for _, arrStr := range arr {
		arrStr = strings.TrimSpace(arrStr)
		if arrStr == "" {
			continue
		}
		arr, err := strconv.Atoi(arrStr)
		if err != nil {
			return res, errorhandler.ErrorBadRequest("invalid string value: " + arrStr)
		}
		arrInt = append(arrInt, int32(arr))
	}
	return arrInt, nil
}
