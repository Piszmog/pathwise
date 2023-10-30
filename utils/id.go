package utils

import "strconv"

func JobRowID(id int) string {
	return "job-" + strconv.Itoa(id) + "-row"
}

func JobRowMetadata(id int) string {
	return "job-" + strconv.Itoa(id) + "-row-metadata"

}

func TimelineStatusRowID(id int) string {
	return "timeline-status-" + strconv.Itoa(id) + "-row"
}

func TimelineStatusRowStringID(id string) string {
	return "timeline-status-" + id + "-row"
}

func TimelineNoteRowID(id int) string {
	return "timeline-note-" + strconv.Itoa(id) + "-row"
}

func TimelineNoteRowStringID(id string) string {
	return "timeline-note-" + id + "-row"
}
