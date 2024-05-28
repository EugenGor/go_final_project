package models

import (
	"fmt"
	"log"
	"time"

	"go_final_project/packages/dateparser"
)

/*
структура Task:
ID - идентификатор задачи
Date - дата задачи в формате "20060102"
Title - заголовок задачи
Comment - комментарий к задаче
Repeat - правило повторения
*/
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// CheckingAndNormalizeDate осуществляет проверку и нормализацию дат
func (t *Task) CheckingAndNormalizeDate() error {
	if t.Title == "" {
		err := fmt.Errorf("the title field is empty")
		return err
	}
	now := time.Now().Truncate(24 * time.Hour)
	log.Printf("today is %v", now)

	if t.Date == "" {
		t.Date = now.Format("20060102")
		log.Println("if t.Date is null.")
		return nil
	}

	if t.Date == "today" {
		t.Date = now.Format("20060102")
		log.Printf("check if %v is equal 'today'", t.Date)
		return nil
	}

	date, err := time.Parse("20060102", t.Date)
	log.Printf("date after parsing: %v", date)
	if err != nil {
		err := fmt.Errorf("the field date is wrong")
		return err
	}

	dt, err := time.Parse("20060102", t.Date)
	if err != nil {
		return err
	}

	// Если дата меньше сегодняшнего числа.
	if now.After(date) {
		//Если правило повторения не указано или равно "", то устанавливается сегодняшнее число
		if t.Repeat == "" {
			log.Printf("repeat rule is empty.")
			t.Date = now.Format("20060102")
		} else {
			log.Printf("repeat rule is not empty.")
			nextDate, err := dateparser.NextDate(now, dt, t.Repeat)
			if err != nil {
				log.Printf("error in NextDate function: %v", err)
				return err
			}
			t.Date = nextDate
		}
	}

	log.Printf("Returning t.Date in TaskCreationRequest function  %v.", t.Date)
	fmt.Println("Error CheckingAndNormalizeDate:", err)
	return nil
}
