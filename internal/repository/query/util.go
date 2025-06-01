package query

import "fmt"

func getIndexedSQLValueParamName(index int, paramName string) string {
	return fmt.Sprintf("%s%d", paramName, index)
}
