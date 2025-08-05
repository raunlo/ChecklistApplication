package domain

import (
	"strings"
)

type SortOrder string

func (s SortOrder) GetValue() string {
	return string(s)
}

func (s *SortOrder) Is(other SortOrder) bool { return *s == other }

func NewSortOrder(value *string) (SortOrder, Error) {
	if value == nil {
		return defaultSortOrder, nil
	}
	valueUpper := strings.ToUpper(*value)
	if (valueUpper != AscSort.GetValue()) && (valueUpper != DescSort.GetValue()) {
		return "", NewError("Value can only be asc or desc", 400)
	}
	return SortOrder(valueUpper), nil
}

const (
	AscSort          SortOrder = "ASC"
	DescSort         SortOrder = "DESC"
	defaultSortOrder           = AscSort
)
