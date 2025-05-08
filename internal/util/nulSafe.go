package util

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
