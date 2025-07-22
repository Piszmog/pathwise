package utils

import "strconv"

func JobRowID(id int64) string {
	return "job-" + strconv.FormatInt(id, 10) + "-row"
}

func JobRowMetadata(id int64) string {
	return "job-" + strconv.FormatInt(id, 10) + "-row-metadata"
}

func TimelineStatusRowID(id int64) string {
	return "timeline-status-" + strconv.FormatInt(id, 10) + "-row"
}

func TimelineStatusRowStringID(id string) string {
	return "timeline-status-" + id + "-row"
}

func TimelineNoteRowID(id int64) string {
	return "timeline-note-" + strconv.FormatInt(id, 10) + "-row"
}

func TimelineNoteRowStringID(id string) string {
	return "timeline-note-" + id + "-row"
}
