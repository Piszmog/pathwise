package types

import (
	"strings"
	"time"
)

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

type JobApplicationTimelineEntry interface {
	Type() JobApplicationTimelineType
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

func (j JobApplicationStatusHistory) Type() JobApplicationTimelineType {
	return JobApplicationTimelineTypeStatus
}

type JobApplicationNote struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	JobApplicationID int
	Note             string
}

func (j JobApplicationNote) Type() JobApplicationTimelineType {
	return JobApplicationTimelineTypeNote
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
