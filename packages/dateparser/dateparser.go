package dateparser

import (
	"time"

	"go_final_project/packages/parser"
)

// Функция NextDate принимает текущее время now, заданную дату date и повторяемость repeat и возвращает следующую дату в формате "20060102".
// Если не удается распарсить значение повторяемости или получить следующую дату, возвращается ошибка.
func NextDate(now time.Time, date time.Time, repeat string) (string, error) {
	pr, err := parser.ParseRepeat(now, date, repeat)
	if err != nil {
		return "", err
	}
	d, err := pr.GetNextDate(now, date)
	if err != nil {
		return "", err
	}
	return d.Format("20060102"), nil
}
