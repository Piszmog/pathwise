package types

type SelectOpts struct {
	Options     []SelectOption
	Err         error
	Name        string
	Label       string
	Placeholder string
	Value       string
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
	TotalApplications           string
	TotalCompanies              string
	AverageTimeToHearBackInDays string
	TotalInterviewingPercentage string
	TotalRejectionsPercentage   string
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
	Page    int
	PerPage int
	Total   int
}

type AlertType string

const (
	AlertTypeError   AlertType = "error"
	AlertTypeSuccess AlertType = "success"
	AlertTypeWarning AlertType = "warning"
)
