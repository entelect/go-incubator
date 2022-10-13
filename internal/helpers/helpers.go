package helpers

// StringSlicesEqual returns true if slice a and b contain the same elements
func StringSlicesEqual(a, b []string, checkOrder bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if checkOrder {
			if v != b[i] {
				return false
			}
		} else {
			if !StringSliceContains(b, v) {
				return false
			}
		}
	}

	return true
}

// StringSliceContains returns true if slice a contains element s
func StringSliceContains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}

	return false
}
