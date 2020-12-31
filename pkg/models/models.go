package models

import "time"

type (
	Post struct {
		PostId       string
		PostMessage  string
		PostImage    []byte
		HashTags     []string
		PostStatus   bool
		PostPriority bool
		Duration	float64
		CreatedOn    time.Time
		UpdatedOn    time.Time
		ScheduleId   string
	}

	PostSchedule struct {
		ScheduleId    string    `json:"schedule_id"`
		ScheduleTitle string    `json:"schedule_title"`
		From          time.Time `json:"from"`
		To            time.Time `json:"to"`
		PostIds       []string  `json:"post_ids"`
		Duration      float64   `json:"duration"`
		CreatedOn     time.Time `json:"created_on"`
		UpdatedOn     time.Time `json:"updated_on"`
	}

	FetchPostResponse struct {
		Data []DbPost `json:"data"`
		Meta Meta     `json:"meta"`
	}

	FetchSchedulePostResponse struct {
		Data []PostSchedule `json:"data"`
		Meta Meta           `json:"meta"`
	}

	StandardResponse struct {
		Data Data `json:"data"`
		Meta Meta `json:"meta"`
	}

	DbPost struct {
		PostId       string    `json:"post_id"`
		PostMessage  string    `json:"post_message"`
		PostImage    string    `json:"post_image"`
		HashTags     []string  `json:"hash_tags"`
		PostStatus   bool      `json:"post_status"`
		PostPriority bool      `json:"post_priority"`
		CreatedOn    time.Time `json:"created_on"`
		UpdatedOn    time.Time `json:"updated_on"`
	}

	Data struct {
		Id        string `json:"id"`
		UiMessage string `json:"ui_message"`
	}

	Meta struct {
		Timestamp     time.Time `json:"timestamp"`
		TransactionId string    `json:"transaction_id"`
		TraceId       string    `json:"trace_id"`
		Status        string    `json:"status"`
	}
)