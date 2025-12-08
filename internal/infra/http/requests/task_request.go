package requests

import (
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
)

type TaskRequest struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description"`
	Deadline    *int64  `json:"deadline"`
}

func (r TaskRequest) ToDomainModel() (interface{}, error) {
	var deadline time.Time
	if r.Deadline != nil {
		if *r.Deadline != 0 {
			deadline = time.Unix(*r.Deadline, 0)
		}
	}

	var dl *time.Time
	if !deadline.IsZero() {
		dl = &deadline
	}
	return domain.Task{
		Title:       r.Title,
		Description: r.Description,
		Deadline:    dl,
	}, nil
}
