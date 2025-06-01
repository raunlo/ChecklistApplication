package dbo

type QueryResult[DBOType interface{} | []interface{}] struct {
	Error  error
	Result DBOType
}
