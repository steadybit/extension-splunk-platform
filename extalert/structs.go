// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

type Response struct {
	Paging  Paging  `json:"paging"`
	Entries []Entry `json:"entry"`
}

type Paging struct {
	Total   int `json:"total"`
	PerPage int `json:"perPage"`
	Offset  int `json:"offset"`
}

type Entry struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Author  string  `json:"author"`
	Content Content `json:"content"`
	Links   Links   `json:"links"`
}

type Content struct {
	Severity    Severity `json:"alert.severity"`
	TriggerTime int64    `json:"trigger_time"`
}

type Links struct {
	Alerts string `json:"alerts"`
}

type Severity int

const (
	SeverityDebug  Severity = 1
	SeverityInfo   Severity = 2
	SeverityWarn   Severity = 3
	SeverityError  Severity = 4
	SeveritySevere Severity = 5
	SeverityFatal  Severity = 6
)

func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "Debug"
	case SeverityInfo:
		return "Info"
	case SeverityWarn:
		return "Warn"
	case SeverityError:
		return "Error"
	case SeveritySevere:
		return "Severe"
	case SeverityFatal:
		return "Fatal"
	default:
		return "Unknown"
	}
}
