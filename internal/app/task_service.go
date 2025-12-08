package app

import (
	"log"
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/database"
)

type TaskService interface {
	Save(t domain.Task) (domain.Task, error)
	FindList(tf database.TasksFilters) ([]domain.Task, error)
	Find(id uint64) (interface{}, error)
	Delete(id uint64) error
	DeleteByTitle(userId uint64, title string) error
	Update(t domain.Task) (domain.Task, error)
	Done(id uint64) (domain.Task, error)
	InProgress(id uint64) (domain.Task, error)
	UpdateDeadline(id uint64, deadline *time.Time) error
}

type taskService struct {
	taskRepo database.TaskRepository
}

func NewTaskService(tr database.TaskRepository) TaskService {
	return taskService{
		taskRepo: tr,
	}
}

func (s taskService) Save(t domain.Task) (domain.Task, error) {
	task, err := s.taskRepo.Save(t)
	if err != nil {
		log.Printf("taskService.Save(s.taskRepo.Save): %s", err)
		return domain.Task{}, err
	}
	return task, nil
}

func (s taskService) FindList(tf database.TasksFilters) ([]domain.Task, error) {
	tasks, err := s.taskRepo.FindList(tf)
	if err != nil {
		log.Printf("taskService.FindList(s.taskRepo.FindList(): %s", err)
		return nil, err
	}
	return tasks, nil
}

func (s taskService) Find(id uint64) (interface{}, error) {
	task, err := s.taskRepo.Find(id)
	if err != nil {
		log.Printf("taskService.Find(s.taskRepo.Find): %s", err)
		return nil, err
	}
	return task, nil
}

func (s taskService) Delete(id uint64) error {
	_, err := s.taskRepo.Find(id)
	if err != nil {
		log.Printf("taskService.Delete(s.taskRepo.Find): %s", err)
		return err
	}

	err = s.taskRepo.Delete(id)
	if err != nil {
		log.Printf("taskService.Delete(s.taskRepo.Delete): %s", err)
		return err
	}

	return nil
}

func (s taskService) DeleteByTitle(userId uint64, title string) error {
	err := s.taskRepo.DeleteByTitle(userId, title)
	if err != nil {
		log.Printf("taskService.DeleteByTitle(s.taskRepo.DeleteByTitle): %s", err)
		return err
	}

	return nil
}

func (s taskService) Update(t domain.Task) (domain.Task, error) {
	updated, err := s.taskRepo.Update(t)
	if err != nil {
		log.Printf("taskService.Update: %s", err)
		return domain.Task{}, err
	}
	return updated, nil
}

func (s taskService) Done(id uint64) (domain.Task, error) {

	_, err := s.taskRepo.Find(id)
	if err != nil {
		return domain.Task{}, err
	}

	task, err := s.taskRepo.Done(id)
	if err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s taskService) InProgress(id uint64) (domain.Task, error) {

	_, err := s.taskRepo.Find(id)
	if err != nil {
		return domain.Task{}, err
	}

	task, err := s.taskRepo.InProgress(id)
	if err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (s taskService) UpdateDeadline(id uint64, deadline *time.Time) error {
	err := s.taskRepo.UpdateDeadline(id, deadline)
	if err != nil {
		log.Printf("taskService.UpdateDeadline: %s", err)
		return err
	}
	return nil
}
