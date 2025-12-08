package database

import (
	"fmt"
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
	"github.com/upper/db/v4"
)

const TasksTableName = "tasks"

type task struct {
	Id          uint64            `db:"id,omitempty"`
	UserId      uint64            `db:"user_id"`
	Title       string            `db:"title"`
	Description *string           `db:"description"`
	Status      domain.TaskStatus `db:"status"`
	Deadline    *time.Time        `db:"deadline"`
	CreatedDate time.Time         `db:"created_date"`
	UpdatedDate time.Time         `db:"updated_date"`
	DeletedDate *time.Time        `db:"deleted_date"`
}

type TaskRepository interface {
	Save(t domain.Task) (domain.Task, error)
	FindList(tf TasksFilters) ([]domain.Task, error)
	Find(id uint64) (domain.Task, error)
	Delete(id uint64) error
	DeleteByTitle(userId uint64, title string) error
	Update(t domain.Task) (domain.Task, error)
	Done(id uint64) (domain.Task, error)
	InProgress(id uint64) (domain.Task, error)
	UpdateDeadline(id uint64, deadline *time.Time) error
}

type taskRepository struct {
	sess db.Session
	coll db.Collection
}

func NewTaskRepository(dbSession db.Session) TaskRepository {
	return taskRepository{
		sess: dbSession,
		coll: dbSession.Collection(TasksTableName),
	}
}

func (r taskRepository) Save(t domain.Task) (domain.Task, error) {
	tsk := r.mapDomainToModel(t)
	tsk.CreatedDate = time.Now()
	tsk.UpdatedDate = time.Now()

	err := r.coll.InsertReturning(&tsk)
	if err != nil {
		return domain.Task{}, err
	}

	t = r.mapModelToDomain(tsk)
	return t, nil
}

func (r taskRepository) Find(id uint64) (domain.Task, error) {
	var t task
	err := r.coll.Find(db.Cond{"id": id, "deleted_date": nil}).One(&t)
	if err != nil {
		return domain.Task{}, err
	}
	return r.mapModelToDomain(t), nil
}

type TasksFilters struct {
	UserId   uint64
	Status   string
	Search   string
	Deadline *time.Time
}

func (r taskRepository) FindList(tf TasksFilters) ([]domain.Task, error) {
	var tasks []task
	query := r.coll.Find(db.Cond{"user_id": tf.UserId, "deleted_date": nil})
	if tf.Status != "" {
		query = query.And(db.Cond{"status": tf.Status})
	}
	if tf.Search != "" {
		query = query.And(db.Cond{"title ilike": "%" + tf.Search + "%"})
	}
	err := query.All(&tasks)
	if err != nil {
		return nil, err
	}
	return r.mapModelToDomainCollection(tasks), nil

}

func (r taskRepository) mapDomainToModel(t domain.Task) task {
	return task{
		Id:          t.Id,
		UserId:      t.UserId,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Deadline:    t.Deadline,
		CreatedDate: t.CreatedDate,
		UpdatedDate: t.UpdatedDate,
		DeletedDate: t.DeletedDate,
	}
}

func (r taskRepository) mapModelToDomain(t task) domain.Task {
	return domain.Task{
		Id:          t.Id,
		UserId:      t.UserId,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Deadline:    t.Deadline,
		CreatedDate: t.CreatedDate,
		UpdatedDate: t.UpdatedDate,
		DeletedDate: t.DeletedDate,
	}
}

func (r taskRepository) mapModelToDomainCollection(ts []task) []domain.Task {
	tasks := make([]domain.Task, len(ts))
	for i, t := range ts {
		tasks[i] = r.mapModelToDomain(t)
	}
	return tasks
}

func (r taskRepository) Delete(id uint64) error {
	var t task
	err := r.coll.Find(db.Cond{"id": id, "deleted_date": nil}).One(&t)
	if err != nil {
		if err == db.ErrNoMoreRows {
			return fmt.Errorf("task with id %d not found or already deleted", id)
		}
		return err
	}

	now := time.Now()
	deleted := &now

	err = r.coll.Find(db.Cond{"id": id}).Update(map[string]interface{}{
		"deleted_date": deleted,
		"updated_date": now,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r taskRepository) DeleteByTitle(userId uint64, title string) error {
	var t task
	err := r.coll.Find(db.Cond{
		"user_id":      userId,
		"title":        title,
		"deleted_date": nil,
	}).One(&t)

	if err != nil {
		if err == db.ErrNoMoreRows {
			return fmt.Errorf("task with title '%s' not found or already deleted", title)
		}
		return err
	}
	now := time.Now()

	err = r.coll.Find(db.Cond{"id": t.Id}).Update(map[string]interface{}{
		"deleted_date": now,
		"updated_date": now,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r taskRepository) Update(t domain.Task) (domain.Task, error) {
	existing, err := r.Find(t.Id)
	if err != nil {
		return domain.Task{}, err
	}

	model := r.mapDomainToModel(existing)

	if t.Title != "" {
		model.Title = t.Title
	}
	if t.Description != nil {
		model.Description = t.Description
	}

	model.UpdatedDate = time.Now()

	err = r.coll.Find(db.Cond{"id": model.Id}).Update(model)
	if err != nil {
		return domain.Task{}, err
	}

	return r.mapModelToDomain(model), nil
}

func (r taskRepository) Done(id uint64) (domain.Task, error) {

	var t task
	err := r.coll.
		Find(db.Cond{"id": id, "deleted_date": nil}).
		One(&t)
	if err != nil {
		return domain.Task{}, err
	}

	now := time.Now()
	err = r.coll.Find(db.Cond{"id": id}).Update(map[string]interface{}{
		"status":       domain.DoneTaskStatus,
		"updated_date": now,
	})
	if err != nil {
		return domain.Task{}, err
	}

	t.Status = domain.DoneTaskStatus
	t.UpdatedDate = now

	return r.mapModelToDomain(t), nil
}

func (r taskRepository) InProgress(id uint64) (domain.Task, error) {

	var t task
	err := r.coll.
		Find(db.Cond{"id": id, "deleted_date": nil}).
		One(&t)
	if err != nil {
		return domain.Task{}, err
	}

	now := time.Now()
	err = r.coll.Find(db.Cond{"id": id}).Update(map[string]interface{}{
		"status":       domain.InProgressTaskStatus,
		"updated_date": now,
	})
	if err != nil {
		return domain.Task{}, err
	}

	t.Status = domain.InProgressTaskStatus
	t.UpdatedDate = now

	return r.mapModelToDomain(t), nil
}

func (r taskRepository) UpdateDeadline(id uint64, deadline *time.Time) error {
	// Перевіряємо чи таск існує і не видалений
	var t task
	err := r.coll.Find(db.Cond{"id": id, "deleted_date": nil}).One(&t)
	if err != nil {
		if err == db.ErrNoMoreRows {
			return fmt.Errorf("task with id %d not found", id)
		}
		return err
	}

	// Готуємо update
	updateFields := map[string]interface{}{
		"updated_date": time.Now(),
	}

	if deadline == nil {
		updateFields["deadline"] = nil
	} else {
		updateFields["deadline"] = *deadline
	}

	// Оновлюємо
	return r.coll.Find(db.Cond{"id": id}).Update(updateFields)
}
