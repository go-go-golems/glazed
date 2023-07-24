package compare

func RemoveDuplicates[T comparable](slice []T) []T {
	encountered := map[T]bool{}
	result := []T{}

	for _, v := range slice {
		if encountered[v] {
			continue
		} else {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return result
}
