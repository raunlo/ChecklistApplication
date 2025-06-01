package util

func StrPointer(s string) *string {
	return &s
}

func UintToPointer(i uint) *uint { return &i }

func Int64Pointer(i int64) *int64 {
	return &i
}

func BoolPointer(boolean bool) *bool {
	return &boolean
}

func AnyPointer[Type any](value Type) *Type {
	return &value
}
