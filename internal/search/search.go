package search

import "time"

type Request struct {
	Title    string   `json:"title,omitempty"`
	Company  string   `json:"company,omitempty"`
	Location string   `json:"location,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	IsRemote bool     `json:"is_remote,omitempty"`
	IsHybrid bool     `json:"is_hybrid,omitempty"`
}

type Response struct {
	JobListings []JobListing `json:"job_listings"`
}

type JobListing struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Location string    `json:"location"`
	IsRemote bool      `json:"is_remote"`
	IsHybrid bool      `json:"is_hybrid"`
	Posted   time.Time `json:"posted"`
}

type Error struct {
	Message string `json:"message"`
}
