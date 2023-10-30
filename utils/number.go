package utils

func GetActualMax(perPage int, total int) int {
	if total < perPage {
		return total
	}
	return perPage
}
