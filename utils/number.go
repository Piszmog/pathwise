package utils

func GetActualMin(start int64, total int64) int64 {
	if total < start {
		return total
	}
	return start
}

func GetActualMax(end int64, total int64) int64 {
	if total < end {
		return total
	}
	return end
}
