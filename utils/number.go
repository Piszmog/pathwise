package utils

func GetActualMin(start int, total int) int {
	if total < start {
		return total
	}
	return start
}

func GetActualMax(end int, total int) int {
	if total < end {
		return total
	}
	return end
}
