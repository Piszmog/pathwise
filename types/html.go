package types

type SelectOpts struct {
	Name        string
	Label       string
	Placeholder string
	Value       string
	Err         error
	Options     []SelectOption
	Required    bool
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
	TotalInterviewingPercentage string
	TotalRejectionsPercentage   string
	TotalApplications           int64
	TotalCompanies              int64
	AverageTimeToHearBackInDays int64
}

type NewTimelineEntry struct {
	Entry   JobApplicationTimelineEntry
	SwapOOB string
}

type FilterOpts struct {
	Company string
	Status  JobApplicationStatus
}

type PaginationOpts struct {
	Page    int64
	PerPage int64
	Total   int64
}

type AlertType string

const (
	AlertTypeError   AlertType = "error"
	AlertTypeSuccess AlertType = "success"
	AlertTypeWarning AlertType = "warning"
)
