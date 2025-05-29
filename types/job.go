package types

import (
	"database/sql"
	"time"
	"unicode"
)

type JobApplication struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	AppliedAt time.Time
	Company   string
	Title     string
	URL       string
	Status    JobApplicationStatus
	ID        int64
	UserID         int64
	Archived       bool
	SalaryMin      sql.NullInt64
	SalaryMax      sql.NullInt64
	SalaryCurrency sql.NullString
}

func (j JobApplication) RecordID() int64 {
	return j.ID
}

type JobApplicationTimelineEntry interface {
	RecordID() int64
	Type() JobApplicationTimelineType
	Created() time.Time
}

type JobApplicationTimelineType string

const (
	JobApplicationTimelineTypeNote   JobApplicationTimelineType = "note"
	JobApplicationTimelineTypeStatus JobApplicationTimelineType = "status"
)

func (t JobApplicationTimelineType) String() string {
	return string(t)
}

func ToJobApplicationTimelineType(val string) JobApplicationTimelineType {
	return jobApplicationTimelineTypeMap[val]
}

var jobApplicationTimelineTypeMap = map[string]JobApplicationTimelineType{
	"note":   JobApplicationTimelineTypeNote,
	"status": JobApplicationTimelineTypeStatus,
}

type JobApplicationStatusHistory struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Status           JobApplicationStatus
	ID               int64
	JobApplicationID int64
}

func (j JobApplicationStatusHistory) RecordID() int64 {
	return j.ID
}

func (j JobApplicationStatusHistory) Type() JobApplicationTimelineType {
	return JobApplicationTimelineTypeStatus
}

func (j JobApplicationStatusHistory) Created() time.Time {
	return j.CreatedAt
}

type JobApplicationNote struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Note             string
	ID               int64
	JobApplicationID int64
}

func (j JobApplicationNote) RecordID() int64 {
	return j.ID
}

func (j JobApplicationNote) Type() JobApplicationTimelineType {
	return JobApplicationTimelineTypeNote
}

func (j JobApplicationNote) Created() time.Time {
	return j.CreatedAt
}

type JobApplicationStatus string

const (
	JobApplicationStatusAccepted     JobApplicationStatus = "accepted"
	JobApplicationStatusApplied      JobApplicationStatus = "applied"
	JobApplicationStatusCanceled     JobApplicationStatus = "canceled"
	JobApplicationStatusClosed       JobApplicationStatus = "closed"
	JobApplicationStatusDeclined     JobApplicationStatus = "declined"
	JobApplicationStatusInterviewing JobApplicationStatus = "interviewing"
	JobApplicationStatusOffered      JobApplicationStatus = "offered"
	JobApplicationStatusRejected     JobApplicationStatus = "rejected"
	JobApplicationStatusWatching     JobApplicationStatus = "watching"
	JobApplicationStatusWithdrawn    JobApplicationStatus = "withdrawn"
)

func (j JobApplicationStatus) String() string {
	return string(j)
}

func (j JobApplicationStatus) PrettyString() string {
	val := j.String()
	if len(val) == 0 {
		return ""
	}
	r := []rune(val)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func ToJobApplicationStatus(val string) JobApplicationStatus {
	return jobApplicationStatusMap[val]
}

var jobApplicationStatusMap = map[string]JobApplicationStatus{
	"accepted":     JobApplicationStatusAccepted,
	"applied":      JobApplicationStatusApplied,
	"canceled":     JobApplicationStatusCanceled,
	"closed":       JobApplicationStatusClosed,
	"declined":     JobApplicationStatusDeclined,
	"interviewing": JobApplicationStatusInterviewing,
	"offered":      JobApplicationStatusOffered,
	"rejected":     JobApplicationStatusRejected,
	"watching":     JobApplicationStatusWatching,
	"withdrawn":    JobApplicationStatusWithdrawn,
}
