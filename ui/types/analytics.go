package types

type SankeyData struct {
	Nodes []SankeyNode `json:"nodes"`
	Links []SankeyLink `json:"links"`
}

type SankeyNode struct {
	Name string `json:"name"`
}

type SankeyLink struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Value  int `json:"value"`
}

type StatusTransition struct {
	FromStatus      string
	ToStatus        string
	TransitionCount int64
}

type StatusCount struct {
	Status string
	Count  int64
}

type AnalyticsData struct {
	SankeyData SankeyData
}
