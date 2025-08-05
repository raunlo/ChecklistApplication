package dbo

type QueryResult[DBOType any | []any] struct {
	Error  error
	Result DBOType
}
