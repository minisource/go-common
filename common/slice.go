package common

func ContainsAll(slice []string, elements []string) bool {
	for _, elem := range elements {
		found := false
		for _, item := range slice {
			if item == elem {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}