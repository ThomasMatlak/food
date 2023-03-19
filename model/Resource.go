package model

import "time"

type Resource struct {
	Created      *time.Time `json:"created"`
	LastModified *time.Time `json:"last_modified"`
	Deleted      *time.Time `json:"deleted"`
}
