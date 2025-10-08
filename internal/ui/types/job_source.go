package types

type JobSource string

const (
	JobSourceHackerNews JobSource = "hackernews"
	JobSourceLinkedIn   JobSource = "linkedin"
	JobSourceIndeed     JobSource = "indeed"
	JobSourceAngelList  JobSource = "angellist"
)

type JobSourceInfo struct {
	Name        string
	DisplayName string
	BadgeClass  string
}

func (js JobSource) Info() JobSourceInfo {
	return jobSourceMap[js]
}

var jobSourceMap = map[JobSource]JobSourceInfo{
	JobSourceHackerNews: {
		Name:        "hackernews",
		DisplayName: "Hacker News",
		BadgeClass:  "bg-orange-50 text-orange-700 ring-1 ring-inset ring-orange-600/20",
	},
	JobSourceLinkedIn: {
		Name:        "linkedin",
		DisplayName: "LinkedIn",
		BadgeClass:  "bg-blue-50 text-blue-700 ring-1 ring-inset ring-blue-600/20",
	},
	JobSourceIndeed: {
		Name:        "indeed",
		DisplayName: "Indeed",
		BadgeClass:  "bg-green-50 text-green-700 ring-1 ring-inset ring-green-600/20",
	},
	JobSourceAngelList: {
		Name:        "angellist",
		DisplayName: "AngelList",
		BadgeClass:  "bg-purple-50 text-purple-700 ring-1 ring-inset ring-purple-600/20",
	},
}
