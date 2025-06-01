package domain

import (
	"strings"
)

type SortOrder string

func (s *SortOrder) GetValue() string {
	return string(*s)
}

func (s *SortOrder) Is(other SortOrder) bool { return *s == other }

func NewSortOrder(value *string) SortOrder {
	return SortOrder(strings.ToUpper(*value))
}

var (
	AscSort  SortOrder = "ASC"
	DescSort SortOrder = "DESC"
)
