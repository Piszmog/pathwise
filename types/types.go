package types

import (
	"time"
	"unicode"
)

type Record interface {
	RecordID() int
}

type JobApplication struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Company   string
	Title     string
	Status    JobApplicationStatus
	URL       string
	AppliedAt time.Time
}

func (j JobApplication) RecordID() int {
	return j.ID
}

type JobApplicationTimelineEntry interface {
	RecordID() int
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
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	JobApplicationID int
	Status           JobApplicationStatus
}

func (j JobApplicationStatusHistory) RecordID() int {
	return j.ID
}

func (j JobApplicationStatusHistory) Type() JobApplicationTimelineType {
	return JobApplicationTimelineTypeStatus
}

func (j JobApplicationStatusHistory) Created() time.Time {
	return j.CreatedAt
}

type JobApplicationNote struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	JobApplicationID int
	Note             string
}

func (j JobApplicationNote) RecordID() int {
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

type SelectOpts struct {
	Name        string
	Label       string
	Placeholder string
	Value       string
	Options     []SelectOption
	Required    bool
	Err         error
}

type SelectOption struct {
	Label string
	Value string
}

var JobApplicationStatusSelectOptions = []SelectOption{
	{
		Label: "Accepted",
		Value: JobApplicationStatusAccepted.String(),
	},
	{
		Label: "Applied",
		Value: JobApplicationStatusApplied.String(),
	},
	{
		Label: "Canceled",
		Value: JobApplicationStatusCanceled.String(),
	},
	{
		Label: "Closed",
		Value: JobApplicationStatusClosed.String(),
	},
	{
		Label: "Declined",
		Value: JobApplicationStatusDeclined.String(),
	},
	{
		Label: "Interviewing",
		Value: JobApplicationStatusInterviewing.String(),
	},
	{
		Label: "Offered",
		Value: JobApplicationStatusOffered.String(),
	},
	{
		Label: "Rejected",
		Value: JobApplicationStatusRejected.String(),
	},
	{
		Label: "Watching",
		Value: JobApplicationStatusWatching.String(),
	},
	{
		Label: "Withdrawn",
		Value: JobApplicationStatusWithdrawn.String(),
	},
}

type StatsOpts struct {
	TotalApplications           string
	TotalCompanies              string
	AverageTimeToHearBackInDays string
	TotalInterviewingPercentage string
	TotalRejectionsPercentage   string
}

type NewTimelineEntry struct {
	SwapOOB string
	Entry   JobApplicationTimelineEntry
}

type FilterOpts struct {
	Company string
	Status  JobApplicationStatus
}

type PaginationOpts struct {
	Page    int
	PerPage int
	Total   int
}
