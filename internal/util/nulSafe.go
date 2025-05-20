package util

import "database/sql"

// NullSafeString safely returns a string from *string
func NullSafeString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// NullSafeInt32 safely returns an int32 from *int32
func NullSafeInt32(i *int32) int32 {
	if i != nil {
		return *i
	}
	return 0
}

func NullableStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func NullableInt32Ptr(ni sql.NullInt32) *int32 {
	if ni.Valid {
		return &ni.Int32
	}
	return nil
}
