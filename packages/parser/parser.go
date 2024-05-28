package parser

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DRepeat хранит число правила d
type DRepeat struct {
	num int
}

/*
ParseDRepeat заполняет структуру DRepeat
Сигнатура правила d: d — задача переносится на указанное число дней.
Не более 1 года/365 дней
*/
func ParseDRepeat(rule []string) (*DRepeat, error) {
	next, err := strconv.Atoi(rule[1])
	if err != nil {
		return nil, fmt.Errorf("error in checking days in repeat rule, got '%s'", rule[1])
	}
	if next > 0 && next <= 365 {
		return &DRepeat{num: next}, nil
	}
	return nil, fmt.Errorf("expected number of days less than 365, got '%s'", rule[1])
}

// GetNextDate вычисляет следующую дату согласно d
// d  — задача переносится на указанное число дней.
func (dr *DRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	result := date
	for {
		result = result.AddDate(0, 0, dr.num)
		if result.After(now) {
			return result, nil
		}
	}
}

type YRepeat struct {
}

/*
ParseYRepeat заполняет структуру YRepeat
y — задача выполняется ежегодно. Этот параметр не требует дополнительных уточнений.
*/
func ParseYRepeat(rule []string) (*YRepeat, error) {
	return &YRepeat{}, nil
}

/*
GetNextDate вычисляет следующую дату по правилу y
При выполнении задачи дата перенесётся на год вперёд
*/
func (yr *YRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	i := 1

	for {
		result := date.AddDate(i, 0, 0)
		if result.After(now) {
			return result, nil
		}
		i++
	}
}

/*
WRepeat хранит список чисел правила w
Сигнатура: w  — задача назначается в указанные дни недели,
где 1 — понедельник, 7 — воскресенье
*/
type WRepeat struct {
	nums []int
}

// ParseWRepeat заполняет структуру WRepeat
func ParseWRepeat(rule []string) (*WRepeat, error) {
	// var week: days of week in rule repeat
	if len(rule) == 1 {
		return nil, fmt.Errorf("error in w rule")
	}

	weekDays := []int{}
	x := strings.Split(rule[1], ",")
	for i := 0; i < len(x); i++ {
		num, err := strconv.Atoi(x[i])
		if err != nil || num > 7 || num < 1 {
			return nil, fmt.Errorf("can not parse days for repeat value")
		}
		weekDays = append(weekDays, num)
	}
	return &WRepeat{nums: weekDays}, nil
}

// GetNextDate вычисляет следующую дату по правилу w. Из списка выбирается ближайший день (по номеру дня в неделе)
func (wr *WRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	startdate := startDateForMWrule(now, date)
	todayWeekday := startdate.Weekday()
	sort.Ints(wr.nums)
	numDay := int(todayWeekday)
	if numDay == 7 {
		numDay = 0
	}

	for _, n := range wr.nums {
		if n > numDay {
			result := startdate.AddDate(0, 0, n-numDay)
			return result, nil
		}
	}

	increment := 7 - int(startdate.Weekday())

	result := startdate.AddDate(0, 0, increment+wr.nums[0])

	return result, nil
}

// MRepeat хранит список дней и список месяцев правила m
type MRepeat struct {
	mDays   []int
	mMonths []int
}

// hasMonths определяет, есть ли в правиле m месяцы
func (mr *MRepeat) hasMonths() bool {
	return len(mr.mMonths) > 0
}

/*
ParseMRepeat принимает на вход массив строк rule, текущее время now и определенную дату date. Она разбирает правило rule для повторения задачи по месяцам и возвращает структуру MRepeat и ошибку.
Сначала функция проверяет длину правила: если оно содержит не 1 и не более 3 элементов, то возвращается ошибка. Затем определяется, содержит ли правило месяцы. Затем создается пустой массив days для хранения дней в правиле.
Далее определяется, от какой даты (now или date) вычислять следующую дату (startdate). Затем происходит разбор дней в правиле: каждый день парсится в целое число, проверяется его корректность и добавляется в массив days.
После этого, если в правиле есть месяцы, то они также разбираются и добавляются в массив months.
В конце создается структура MRepeat с массивами дней и месяцев, и эта структура возвращается вместе с ошибкой (если она есть).
*/
func ParseMRepeat(rule []string, now time.Time, date time.Time) (*MRepeat, error) {

	if len(rule) == 1 || len(rule) > 3 {
		return nil, fmt.Errorf("error in m rule")
	}
	hasMonths := false

	if len(rule) == 3 {
		hasMonths = true
	}

	days := []int{}

	daysInRule := strings.Split(rule[1], ",")

	startdate := startDateForMWrule(now, date)

	for _, day := range daysInRule {
		num, err := strconv.Atoi(day)
		if err != nil {
			return nil, fmt.Errorf("error in checking days in repeat rule 'm', got '%s'", day)
		}
		if num >= 1 && num <= 31 {
			days = append(days, num)
		} else if num == -1 {
			t := Date(startdate.Year(), int(startdate.Month()+1), 0)
			days = append(days, int(t.Day()))
		} else if num == -2 {
			t := Date(startdate.Year(), int(startdate.Month()+1), 0)
			days = append(days, int(t.Day())-1)
		} else {
			return nil, fmt.Errorf("error in checking days in repeat rule 'm', got '%s'", day)
		}

	}

	months := []int{}

	if hasMonths {
		monthsInRule := strings.Split(rule[2], ",")

		for _, month := range monthsInRule {
			num, err := strconv.Atoi(month)
			if err != nil || num < 1 || num > 12 {
				return nil, fmt.Errorf("error in checking days in repeat rule 'm', got '%s'", month)
			}
			months = append(months, num)
		}
	}
	return &MRepeat{mDays: days, mMonths: months}, nil
}

// Функция GetNextDate принимает на вход текущее время now и определенную дату date и возвращает следующую дату, соответствующую правилу повторения MRepeat
func (mr *MRepeat) GetNextDate(now time.Time, date time.Time) (time.Time, error) {
	startdate := startDateForMWrule(now, date) //определяет начальную дату

	sort.Ints(mr.mDays) //mr.mDays сортируется по возрастанию

	var nextDay time.Time
	/*
		Далее функция проверяет, есть ли месяцы в правиле.
		Если месяцы отсутствуют, то происходит перебор дней в массиве mr.mDays.
		Если найденный день больше текущего дня (startdate.Day()), то вычисляется следующая дата (nextDay) путем добавления разницы между найденным днем и текущим днем к startdate.
		Если полученная дата (nextDay) не совпадает с ожидаемым днем, то вычисляется следующая дата путем создания новой даты с тем же годом и следующим месяцем, но указанным днем.
	*/
	if !mr.hasMonths() {
		for _, day := range mr.mDays {
			if day > int(startdate.Day()) {
				nextDay = startdate.AddDate(0, 0, day-int(startdate.Day()))
				if nextDay.Day() != day {
					nextDay = Date(startdate.Year(), int(startdate.Month())+1, day)
				}
				return nextDay, nil
			}
		}
		/*
			Если в массиве mr.mDays нет подходящего дня, то функция проверяет, есть ли у правила "нулевая" дата (0001-01-01).
			Если есть, то начальная дата (startdate) устанавливается на первое число следующего месяца и выполняется тот же перебор дней в массиве mr.mDays.
			Если найденный день больше или равен текущему дню (startdate.Day()), то вычисляется следующая дата (nextDay) путем добавления разницы между найденным днем и текущим днем к startdate.
		*/
		if nextDay == Date(0001, 1, 1) {
			startdate = Date(int(startdate.Year()), int(startdate.Month())+1, 1)
			for _, day := range mr.mDays {
				if day >= int(startdate.Day()) {
					nextDay = startdate.AddDate(0, 0, day-int(startdate.Day()))
					return nextDay, nil
				}
			}
		}
	}

	if mr.hasMonths() {
		sort.Ints(mr.mMonths)

		nextDay = ruleMwithMonth(startdate, mr.mDays, mr.mMonths)
		return nextDay, nil
	}

	return time.Time{}, fmt.Errorf("error in checking days and months in 'm' repeat rule")
}

type RepeatRule interface {
	GetNextDate(now time.Time, date time.Time) (time.Time, error)
}

// Date возвращает тип времени из int-типов year, month и day
func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// startDateForMWrule помогает определить начальную точку для вычисления следующей даты в соответствии с правилом повторения, учитывая как текущее время, так и указанную дату.
func startDateForMWrule(now time.Time, date time.Time) time.Time {
	if date.After(now) {
		return date
	}
	return now
}

/*
ruleMwithMonth возвращает следующую дату, удовлетворяющую заданным условиям, или нулевое значение времени, если такая дата не найдена.
Выполняются следующие действия:
1. Для каждого месяца из массива mMonths проверяется, равен ли он текущему месяцу начальной даты startdate.
2. Если месяц совпадает с текущим месяцем начальной даты, то начальная дата устанавливается на первое число этого месяца.
3. Затем определяется количество дней в этом месяце.
4. Для каждого дня из массива mDays проверяется, больше ли он текущего дня начальной даты и не превышает ли он количество дней в месяце.
5. Если условия выполняются, то устанавливается следующая дата и возвращается.
6. Если день больше текущего дня, но превышает количество дней в месяце, то начальная дата устанавливается на первое число следующего месяца.
7. Если месяц из массива mMonths больше текущего месяца начальной даты, то аналогично выполняются шаги с 2 по 6 для этого месяца.
*/
func ruleMwithMonth(startdate time.Time, mDays []int, mMonths []int) time.Time {
	var nextDay time.Time

	for _, month := range mMonths {
		if month == int(startdate.Month()) {
			startdate = Date(startdate.Year(), month, 1)

			t := Date(startdate.Year(), int(startdate.Month())+1, 0)
			dayInMonth := t.Day()

			for _, day := range mDays {
				if day > int(startdate.Day()) && day <= dayInMonth {
					gotDay := Date(startdate.Year(), int(startdate.Month()), day)
					nextDay = gotDay
					return nextDay
				} else if day > int(startdate.Day()) && day > dayInMonth {
					startdate = Date(startdate.Year(), int(startdate.Month())+1, 1)
				}
			}
		} else if month > int(startdate.Month()) {

			startdate = Date(startdate.Year(), month, 1)

			t := Date(startdate.Year(), int(startdate.Month())+1, 0)
			dayInMonth := t.Day()

			for _, day := range mDays {
				if day >= int(startdate.Day()) && day <= dayInMonth {
					gotDay := Date(startdate.Year(), int(startdate.Month()), day)
					nextDay = gotDay
					return nextDay
				} else if day > int(startdate.Day()) && day > dayInMonth {
					startdate = Date(startdate.Year(), int(startdate.Month())+1, 1)
				}
			}
		}
	}
	return nextDay
}

/*
ParseRepeat выполняет следующие действия:
1. Если строка repeat пустая, то функция возвращает ошибку.
2. В противном случае, она разбивает строку repeat на отдельные части с помощью пробелов и сохраняет их в массив rule.
3. Далее, функция анализирует первую часть массива rule с помощью оператора switch.
4. В зависимости от значения первой части массива rule, функция вызывает соответствующую функцию для разбора правила повторения (ParseYRepeat, ParseDRepeat, ParseWRepeat или ParseMRepeat).
5. Если происходит ошибка при разборе правила повторения, то функция возвращает эту ошибку.
6. Если тип повторения неизвестен, то функция возвращает ошибку.
*/
func ParseRepeat(now time.Time, date time.Time, repeat string) (RepeatRule, error) {
	if repeat == "" {
		return nil, fmt.Errorf("expected repeat, got an empty string")
	}

	rule := strings.Split(repeat, " ")

	var parsedRepeat RepeatRule
	var err error

	log.Printf("rule[0] before switch is: %v", rule[0])
	switch {
	case rule[0] == "y":
		parsedRepeat, err = ParseYRepeat(rule)
		if err != nil {
			return nil, err
		}
	case rule[0] == "d":
		parsedRepeat, err = ParseDRepeat(rule)
		if err != nil {
			return nil, err
		}
	case rule[0] == "w":
		parsedRepeat, err = ParseWRepeat(rule)
		if err != nil {
			return nil, err
		}
	case rule[0] == "m":
		parsedRepeat, err = ParseMRepeat(rule, now, date)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unkown repeat identifier %s", rule[0])
	}

	return parsedRepeat, nil
}
