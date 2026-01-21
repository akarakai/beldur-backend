package httperr

import (
	"errors"
	"net/http"
	"time"
)

type Response struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Mapped struct {
	Status  int
	Code    string
	Message string
}

type Manager struct {
	defaultMapped Mapped
	mappings      []entry
}

type entry struct {
	target error
	m      Mapped
}

func NewManager() *Manager {
	return &Manager{
		defaultMapped: Mapped{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "Internal server error",
		},
	}
}

func (m *Manager) Add(target error, mapped Mapped) {
	m.mappings = append(m.mappings, entry{target: target, m: mapped})
}

func (m *Manager) Map(err error) (status int, resp Response) {
	mapped := m.defaultMapped
	for _, e := range m.mappings {
		if errors.Is(err, e.target) {
			mapped = e.m
			break
		}
	}

	return mapped.Status, Response{
		Code:      mapped.Code,
		Message:   mapped.Message,
		Timestamp: time.Now(),
	}
}
