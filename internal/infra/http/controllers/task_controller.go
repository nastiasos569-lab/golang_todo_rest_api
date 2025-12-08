package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/app"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/database"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/http/requests"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/http/resources"
)

type TaskController struct {
	taskService app.TaskService
}

func NewTaskController(ts app.TaskService) TaskController {
	return TaskController{
		taskService: ts,
	}
}

func (c TaskController) Save() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task, err := requests.Bind(r, requests.TaskRequest{}, domain.Task{})
		if err != nil {
			log.Printf("TaskController.Save(requests.Bind):%s", err)
			BadRequest(w, err)
			return
		}
		task.UserId = user.Id
		task.Status = domain.NewTaskStatus
		task, err = c.taskService.Save(task)
		if err != nil {
			log.Printf("TaskController.Save(requests.Bind):%s", err)
			InternalServerError(w, err)
			return
		}
		taskDto := resources.TaskDto{}
		taskDto = taskDto.DomainToDto(task)
		Success(w, taskDto)
	}
}
func (c TaskController) Find() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task := r.Context().Value(TaskKey).(domain.Task)

		if user.Id != task.UserId {
			err := errors.New("access denied")
			Forbidden(w, err)
			return
		}

		taskDto := resources.TaskDto{}
		taskDto = taskDto.DomainToDto(task)
		Success(w, taskDto)
	}
}

func (c TaskController) FindList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)

		status := ""
		if r.URL.Query().Has("status") {
			status = r.URL.Query().Get("status")
		}

		search := ""
		if r.URL.Query().Has("search") {
			search = r.URL.Query().Get("search")
		}

		filters := database.TasksFilters{
			UserId: user.Id,
			Status: status,
			Search: search,
		}

		tasks, err := c.taskService.FindList(filters)
		if err != nil {
			log.Printf("TaskController.FindList(c.taskService.FindList))")
			InternalServerError(w, err)
			return
		}
		tasksDto := resources.TasksDto{}
		tasksDto = tasksDto.DomainToDto(tasks)
		Success(w, tasksDto)
	}
}

func (c TaskController) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task := r.Context().Value(TaskKey).(domain.Task)

		if user.Id != task.UserId {
			err := errors.New("access denied")
			Forbidden(w, err)
			return
		}

		err := c.taskService.Delete(task.Id)
		if err != nil {
			log.Printf("TaskController.Delete(c.taskService.Delete): %s", err)
			InternalServerError(w, err)
			return
		}

		Success(w, map[string]string{"message": "task deleted"})
	}
}

func (c TaskController) DeleteByTitle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)

		title := r.URL.Query().Get("title")
		if title == "" {
			err := errors.New("missing 'title' parameter")
			BadRequest(w, err)
			return
		}

		err := c.taskService.DeleteByTitle(user.Id, title)
		if err != nil {
			log.Printf("TaskController.DeleteByTitle(c.taskService.DeleteByTitle): %s", err)
			InternalServerError(w, err)
			return
		}

		Success(w, map[string]string{"message": "task deleted by title"})
	}
}

func (c TaskController) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task := r.Context().Value(TaskKey).(domain.Task)

		if user.Id != task.UserId {
			Forbidden(w, errors.New("access denied"))
			return
		}

		req, err := requests.Bind(r, requests.TaskRequest{}, domain.Task{})
		if err != nil {
			BadRequest(w, err)
			return
		}

		req.Id = task.Id
		req.UserId = user.Id

		updated, err := c.taskService.Update(req)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		dto := resources.TaskDto{}.DomainToDto(updated)
		Success(w, dto)
	}
}

func (c TaskController) Done() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task := r.Context().Value(TaskKey).(domain.Task)

		if task.UserId != user.Id {
			Forbidden(w, errors.New("access denied"))
			return
		}

		updated, err := c.taskService.Done(task.Id)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		dto := resources.TaskDto{}.DomainToDto(updated)
		Success(w, dto)
	}
}

func (c TaskController) InProgress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task := r.Context().Value(TaskKey).(domain.Task)

		if task.UserId != user.Id {
			Forbidden(w, errors.New("access denied"))
			return
		}

		updated, err := c.taskService.InProgress(task.Id)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		dto := resources.TaskDto{}.DomainToDto(updated)
		Success(w, dto)
	}
}

func (c TaskController) UpdateDeadline() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		task := r.Context().Value(TaskKey).(domain.Task)

		if user.Id != task.UserId {
			Forbidden(w, errors.New("access denied"))
			return
		}

		req := struct {
			Deadline *int64 `json:"deadline"`
		}{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, err)
			return
		}

		var dl *time.Time
		if req.Deadline != nil {
			if *req.Deadline != 0 {
				t := time.Unix(*req.Deadline, 0)
				dl = &t
			}
		}

		err := c.taskService.UpdateDeadline(task.Id, dl)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		updated, _ := c.taskService.Find(task.Id)
		dto := resources.TaskDto{}.DomainToDto(updated.(domain.Task))
		Success(w, dto)
	}
}
