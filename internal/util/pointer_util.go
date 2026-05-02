package util

//go:fix inline
func StrPointer(s string) *string {
	return new(s)
}

//go:fix inline
func UintToPointer(i uint) *uint { return new(i) }

//go:fix inline
func Int64Pointer(i int64) *int64 {
	return new(i)
}

//go:fix inline
func BoolPointer(boolean bool) *bool {
	return new(boolean)
}

//go:fix inline
func AnyPointer[Type any](value Type) *Type {
	return new(value)
}
