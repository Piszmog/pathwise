package types

import (
	"strings"
	"time"
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
	JobApplicationTimelineTypeNote   JobApplicationTimelineType = "Note"
	JobApplicationTimelineTypeStatus JobApplicationTimelineType = "Status"
)

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
	JobApplicationStatusAccepted     JobApplicationStatus = "Accepted"
	JobApplicationStatusApplied      JobApplicationStatus = "Applied"
	JobApplicationStatusCanceled     JobApplicationStatus = "Canceled"
	JobApplicationStatusClosed       JobApplicationStatus = "Closed"
	JobApplicationStatusDeclined     JobApplicationStatus = "Declined"
	JobApplicationStatusInterviewing JobApplicationStatus = "Interviewing"
	JobApplicationStatusOffered      JobApplicationStatus = "Offered"
	JobApplicationStatusRejected     JobApplicationStatus = "Rejected"
	JobApplicationStatusWatching     JobApplicationStatus = "Watching"
	JobApplicationStatusWithdrawn    JobApplicationStatus = "Withdrawn"
)

func (j JobApplicationStatus) String() string {
	return string(j)
}

func ToJobApplicationStatus(val string) JobApplicationStatus {
	switch strings.ToLower(val) {
	case strings.ToLower(JobApplicationStatusAccepted.String()):
		return JobApplicationStatusAccepted
	case strings.ToLower(JobApplicationStatusApplied.String()):
		return JobApplicationStatusApplied
	case strings.ToLower(JobApplicationStatusCanceled.String()):
		return JobApplicationStatusCanceled
	case strings.ToLower(JobApplicationStatusClosed.String()):
		return JobApplicationStatusClosed
	case strings.ToLower(JobApplicationStatusDeclined.String()):
		return JobApplicationStatusDeclined
	case strings.ToLower(JobApplicationStatusInterviewing.String()):
		return JobApplicationStatusInterviewing
	case strings.ToLower(JobApplicationStatusOffered.String()):
		return JobApplicationStatusOffered
	case strings.ToLower(JobApplicationStatusRejected.String()):
		return JobApplicationStatusRejected
	case strings.ToLower(JobApplicationStatusWatching.String()):
		return JobApplicationStatusWatching
	case strings.ToLower(JobApplicationStatusWithdrawn.String()):
		return JobApplicationStatusWithdrawn
	default:
		return JobApplicationStatusWatching
	}
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
		Value: strings.ToLower(JobApplicationStatusAccepted.String()),
	},
	{
		Label: "Applied",
		Value: strings.ToLower(JobApplicationStatusApplied.String()),
	},
	{
		Label: "Canceled",
		Value: strings.ToLower(JobApplicationStatusCanceled.String()),
	},
	{
		Label: "Closed",
		Value: strings.ToLower(JobApplicationStatusClosed.String()),
	},
	{
		Label: "Declined",
		Value: strings.ToLower(JobApplicationStatusDeclined.String()),
	},
	{
		Label: "Interviewing",
		Value: strings.ToLower(JobApplicationStatusInterviewing.String()),
	},
	{
		Label: "Offered",
		Value: strings.ToLower(JobApplicationStatusOffered.String()),
	},
	{
		Label: "Rejected",
		Value: strings.ToLower(JobApplicationStatusRejected.String()),
	},
	{
		Label: "Watching",
		Value: strings.ToLower(JobApplicationStatusWatching.String()),
	},
	{
		Label: "Withdrawn",
		Value: strings.ToLower(JobApplicationStatusWithdrawn.String()),
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
	First   bool
}
