package utils

import (
	"github.com/Piszmog/pathwise/types"
)

func GetFirstElementID[T types.Record](vals []T) int {
	if len(vals) == 0 {
		return 0
	}
	return vals[0].RecordID()
}

func GetFirstElementType(vals []types.JobApplicationTimelineEntry) types.JobApplicationTimelineType {
	if len(vals) == 0 {
		return ""
	}
	return vals[0].Type()
}
