package tasks_repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go_final_project/packages/dateparser"
	"go_final_project/packages/models"
)

const (
	limitConst = 20
)

// структура для работы с Tasks
type TasksRepository struct {
	db *sql.DB
}

func NewTasksRepository(db *sql.DB) TasksRepository {
	return TasksRepository{db: db}
}

func (tr TasksRepository) AddTask(t models.Task) (int, error) {
	task, err := tr.db.Exec("insert into scheduler (date, title, comment, repeat) values (:date, :title, :comment, :repeat)",
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat))

	if err != nil {
		return 0, fmt.Errorf("failed to add task to the database: %w", err)
	}

	id, err := task.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// PostTaskDone перемещает задачу в соответствии с правилом повторения
func (tr TasksRepository) PostTaskDone(id int) (*models.Task, error) {
	t, err := tr.GetTask(id)
	if err != nil {
		return nil, err
	}

	dt, err := time.Parse("20060102", t.Date)
	if err != nil {
		return nil, err
	}

	if t.Repeat == "" {
		fmt.Println("Repeat is null")
		err = tr.DeleteTask(id)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	fmt.Println("Repeat is not null")
	now := time.Now()
	nextDate, err := dateparser.NextDate(now, dt, t.Repeat)
	if err != nil {
		return nil, err
	}
	err = tr.UpdateTaskDate(t, nextDate)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// UpdateTaskInBd обновляет задачу в базе данных.
func (tr TasksRepository) UpdateTaskInBd(t models.Task) error {
	_, err := tr.db.Exec("update scheduler set date = :date, title = :title, comment = :comment,"+
		"repeat = :repeat WHERE id = :id",
		sql.Named("date", t.Date),
		sql.Named("title", t.Title),
		sql.Named("comment", t.Comment),
		sql.Named("repeat", t.Repeat),
		sql.Named("id", t.ID))

	if err != nil {
		return err
	}

	return nil
}

// GetTask возвращает только одну строку с определенным id
func (tr TasksRepository) GetTask(id int) (models.Task, error) {
	s := models.Task{}
	row := tr.db.QueryRow("select * from scheduler where id = :id",
		sql.Named("id", id))

	// заполняем объект TaskCreationRequest данными из таблицы
	err := row.Scan(&s.ID, &s.Date, &s.Title, &s.Comment, &s.Repeat)
	if err != nil {
		return s, err
	}
	return s, nil
}

// Возвращаем сроки с ближайшими датами.
func (tr TasksRepository) GetAllTasks() ([]models.Task, error) {
	today := time.Now().Format("20060102")

	rows, err := tr.db.Query("select * from scheduler where date >= :today "+
		"order by date limit :limit",
		sql.Named("today", today),
		sql.Named("limit", limitConst))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := []models.Task{}
	// заполняем объект Task данными из таблицы
	for rows.Next() { // пока есть записи
		s := models.Task{} // создаем новый объект  Task и заполняем его данными из текущего row
		err := rows.Scan(&s.ID, &s.Date, &s.Title, &s.Comment, &s.Repeat)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type DateSearchParam struct {
	Date time.Time
}

func (dp *DateSearchParam) GetQueryData() *QueryData {
	return &QueryData{
		Param:     dp.Date.Format("20060102"),
		Condition: "where date like :search",
	}
}

type TextSearchParam struct {
	Text string
}

func (tp *TextSearchParam) GetQueryData() *QueryData {
	return &QueryData{
		Param:     fmt.Sprintf("%%%s%%", tp.Text),
		Condition: "where title like :search or comment like :search",
	}
}

type SearchQueryData interface {
	GetQueryData() *QueryData
}

func QueryDataFromString(search string) SearchQueryData {
	searchDate, err := time.Parse("02.01.2006", search)
	if err != nil {
		return &TextSearchParam{Text: search}
	} else {
		return &DateSearchParam{Date: searchDate}
	}
}

type QueryData struct {
	Param     string
	Condition string
}

// SearchTasks вернет строки в соответсвии с критерием поиска search
func (tr TasksRepository) SearchTasks(searchData SearchQueryData) ([]models.Task, error) {
	limitConst := 20
	var rows *sql.Rows

	queryData := searchData.GetQueryData()

	querySQL := strings.Join([]string{
		"select id, date, title, comment, repeat from scheduler",
		queryData.Condition,
		"order by date limit :limit",
	}, " ")

	rows, err := tr.db.Query(querySQL,
		sql.Named("search", queryData.Param),
		sql.Named("limit", limitConst))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := []models.Task{}
	// заполняем Task данными из таблицы
	for rows.Next() {
		// создаем новый объект Task и заполняем его данными из текущей итерации
		s := models.Task{}
		if err := rows.Scan(&s.ID, &s.Date, &s.Title, &s.Comment, &s.Repeat); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteTask удаляет запись согласно id
func (tr TasksRepository) DeleteTask(id int) error {
	_, err := tr.db.Exec("delete from scheduler where id = :id",
		sql.Named("id", id))
	if err != nil {
		return err
	}

	return nil
}

// UpdateTaskInBd обновляет задачу в базе данных согласно обновленной дате
func (tr TasksRepository) UpdateTaskDate(t models.Task, newDate string) error {
	_, err := tr.db.Exec("update scheduler set date = :date where id = :id",
		sql.Named("date", newDate),
		sql.Named("id", t.ID))

	if err != nil {
		return err
	}

	return nil
}
