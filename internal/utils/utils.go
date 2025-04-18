package utils

// DerefString returns the value of a string pointer or an empty string if the pointer is nil
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefBool returns the value of a bool pointer or false if the pointer is nil
func DerefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// DerefInt returns the value of an int pointer or 0 if the pointer is nil
func DerefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
