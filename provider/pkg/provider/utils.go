package provider

import "sort"

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func InSlice(a string, b []string) bool {
	for _, v := range b {
		if v == a {
			return true
		}
	}
	return false
}
