package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go_final_project/packages/config"
	"go_final_project/packages/dateparser"
	"go_final_project/packages/models"
	"go_final_project/packages/repp"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

const (
	MarshallingError    = "error in marshalling JSON"
	UnMarshallingError  = "error in unmarshalling JSON"
	ResponseWriteError  = "error in writing data"
	ReadingError        = "error in reading data"
	InvalidIdError      = "invalid id"
	IdMissingError      = "id is missing"
	InvalidDateError    = "invalid date"
	InvalidNowDateError = "invalid now date"
	InvalidRepeatError  = "invalid repeat value"
	InternalServerError = "internal server error"
	ValidatingDateError = "error in validating date"
)

// repeatRulePattern проверяет правило repeat
var repeatRulePattern *regexp.Regexp = regexp.MustCompile(`^([mwd]\s\S.*|y$)`)

// структуры для маршалинга в Json
type signinRequest struct {
	Password string `json:"password"`
}

type signinResponse struct {
	Token string `json:"token"`
}

type apiError struct {
	Error string `json:"error"`
}

func NewApiError(err error) apiError {
	return apiError{Error: err.Error()}
}

func (e apiError) ToJson() ([]byte, error) {
	res, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func RenderApiError(w http.ResponseWriter, err error, status int) {
	apiErr := NewApiError(err)
	errorJson, _ := apiErr.ToJson()
	http.Error(w, string(errorJson), status)
}

// GetNextDay ищет следующий день для задачи
func GetNextDay(w http.ResponseWriter, r *http.Request) {
	//извлекаем значения now, date и repeat из параметров URL-запроса HTTP-запроса.
	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	//проверка параметров date, now и repeat, чтобы убедиться, что они имеют правильный формат и содержат допустимые значения.
	//Если какая-либо из проверок завершается ошибкой, возвращается ответ об ошибке с соответствующим сообщением/const об ошибке.
	//date
	_, err := strconv.Atoi(date)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidDateError), http.StatusBadRequest)
		return
	}

	dtParsed, err := time.Parse("20060102", date)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidDateError), http.StatusBadRequest)
		return
	}

	//now
	_, err = strconv.Atoi(now)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidNowDateError), http.StatusBadRequest)
		return
	}

	dtNow, err := time.Parse("20060102", now)
	if err != nil {
		err := fmt.Errorf("incorrect date: %v", err)
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidNowDateError), http.StatusBadRequest)
		return
	}

	//repeat
	if repeat == "" {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidRepeatError), http.StatusBadRequest)
		return
	} else if !repeatRulePattern.MatchString(repeat) {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidRepeatError), http.StatusBadRequest)
		return
	}
	//получаем следующий день
	nextDay, err := dateparser.NextDate(dtNow, dtParsed, repeat)

	if err != nil {
		err := fmt.Errorf("incorrect repetition value")
		w.WriteHeader(http.StatusInternalServerError)
		WriteResponse(w, []byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	WriteResponse(w, []byte(nextDay)) // w.Write([]byte(nextDay))
}

// WriteResponse при ошибке вызывает panic/fatal
func WriteResponse(w http.ResponseWriter, s []byte) {
	_, err := w.Write(s)
	if err != nil {
		log.Fatalf("Can not write response: %s", err.Error())
	}
}

// Api структурой для работы с API
type Api struct {
	repp   *repp.TasksRepository
	config *config.Config
}

// NewApi конструктор объекта api.
func NewApi(repp *repp.TasksRepository, config *config.Config) *Api {
	// создаем ссылку на объект api
	return &Api{repp: repp, config: config}
}

/*
В зависимости от метода HTTP-запроса (GET, POST, PUT, DELETE), TaskHandler выполняет различные действия:
1. Если метод запроса - GET, TaskHandler извлекает параметр "id" из URL-запроса, преобразует его в целое число и вызывает метод GetTask объекта Api для получения информации о задаче с указанным идентификатором.
2. Если метод запроса - POST, TaskHandler вызывает метод CreateTask объекта Api для создания новой задачи.
3. Если метод запроса - PUT, TaskHandler вызывает метод UpdateTask объекта Api для обновления существующей задачи.
4. Если метод запроса - DELETE, TaskHandler проверяет наличие параметра "id" в URL-запросе. Если он присутствует, то вызывается метод DeleteTask объекта Api для удаления задачи с указанным идентификатором. В противном случае возвращается ошибка "IdMissingError".
*/
func (a *Api) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		idToSearch := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idToSearch)
		if err != nil {
			log.Println("error:", err)
			RenderApiError(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
			return
		}
		a.GetTask(w, r, id)

	case r.Method == http.MethodPost:
		a.CreateTask(w, r)
	case r.Method == http.MethodPut:
		a.UpdateTask(w, r)
	case r.Method == http.MethodDelete:
		idToSearch := r.URL.Query().Get("id")
		if idToSearch != "" {
			a.DeleteTask(w, r)
		} else {
			RenderApiError(w, fmt.Errorf(IdMissingError), http.StatusBadRequest)
			return
		}
	}
}

// извлечение id
func (a *Api) GetTaskByIdHandler(w http.ResponseWriter, r *http.Request) {
	idToSearch := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(IdMissingError), http.StatusBadRequest)
		return
	}
	a.GetTask(w, r, id)
}

// проверяем search
func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("search") != "" {
		s := r.URL.Query().Get("search")
		a.SearchTasks(w, r, s)
	} else {
		a.GetAllTasks(w)
	}
}

// выполняем поиск согласно критерию и отправляем клиенту json
func (a *Api) GetAllTasks(w http.ResponseWriter) {
	foundTasks, err := a.repp.GetAllTasks()
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	result := make(map[string][]models.Task)
	result["tasks"] = foundTasks

	resp, err := json.Marshal(result)
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(ResponseWriteError), http.StatusBadRequest)
		return
	}
}

// ищем задачу согласно search из GetTasksHandler
func (a *Api) SearchTasks(w http.ResponseWriter, r *http.Request, search string) {
	foundTasks, err := a.repp.SearchTasks(repp.QueryDataFromString(search))
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	result := make(map[string][]models.Task)
	result["tasks"] = foundTasks

	resp, err := json.Marshal(result)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(resp)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(ResponseWriteError), http.StatusBadRequest)
		return
	}
}

// отправляем задачу в базу данных
func (a *Api) CreateTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(ReadingError), http.StatusBadRequest)
		return
	}
	log.Println("received:", buf.String())

	parseBody := models.Task{}
	err = json.Unmarshal(buf.Bytes(), &parseBody)
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(UnMarshallingError), http.StatusBadRequest)
		return
	}

	err = parseBody.CheckingAndNormalizeDate()
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(ValidatingDateError), http.StatusBadRequest)
		return
	}

	id, err := a.repp.AddTask(parseBody)
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	WriteResponse(w, []byte(fmt.Sprintf("{\"id\":%d}", id))) //
}

// обновляем задачу в базе данных
func (a *Api) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(ReadingError), http.StatusBadRequest)
		return
	}

	parseBody := models.Task{}
	err = json.Unmarshal(buf.Bytes(), &parseBody)
	if err != nil {
		log.Println("err:", err)
		RenderApiError(w, fmt.Errorf(UnMarshallingError), http.StatusBadRequest)
		return
	}

	err = parseBody.CheckingAndNormalizeDate()
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(ValidatingDateError), http.StatusBadRequest)
		return
	}
	idToSearch, err := strconv.Atoi(parseBody.ID)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	_, err = a.repp.GetTask(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	err = a.repp.UpdateTaskInBd(parseBody)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	jsonItem, err := json.Marshal(parseBody)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	WriteResponse(w, jsonItem)
}

/*
ищем задачу согласно полученому id. При ошибке возвращаем ошибку с кодом 400
удаляем задачу согласно id. При ошибке возвращаем ошибку с кодом 500
при успешном удалении возвращаем статус 200 и записываем пустой объект JSON в тело ответа с помощью WriteResponse(w, []byte("{}"))
*/
func (a *Api) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idToSearch := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	err = a.repp.DeleteTask(id)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidIdError), http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		WriteResponse(w, []byte("{}"))
		return
	}
}

/*
TaskDoneHandler получает значение параметра id из URL-запроса с помощью r.URL.Query().Get("id"). Затем преобразует его в целое число с помощью strconv.Atoi(idToSearch).
Если происходит ошибка при преобразовании, логирует ошибку, возвращаем ошибку с кодом 400 и сообщением InvalidIdError.
Затем вызывает метод PostTaskDone у a.repp, передавая ему полученное id. Если возвращенная задача newTask равна nil, устанавливает статус ответа 200 (OK) и записываем пустой объект JSON в тело ответа с помощью WriteResponse(w, []byte("{}")).
Если при выполнении метода PostTaskDone возникла ошибка, логирует её, TaskDoneHandler возвращаем ошибку с кодом 500 и сообщением InternalServerError.
Если операция прошла успешно (без ошибок), устанавливает статус ответа 200 (OK) и записывает пустой объект JSON в тело ответа с помощью WriteResponse(w, []byte("{}"))
*/
func (a *Api) TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	idToSearch := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idToSearch)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InvalidIdError), http.StatusBadRequest)
		return
	}

	newTask, err := a.repp.PostTaskDone(id)
	if newTask == nil {
		w.WriteHeader(http.StatusOK)
		WriteResponse(w, []byte("{}"))
		return
	}
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		WriteResponse(w, []byte("{}"))
		return
	}
}

/*
GetTask вызывает метод GetTask у a.repp, передавая ему полученное значение id.
Если при выполнении метода GetTask возникает ошибка, мы логируем её, устанавливаем статус ответа 500 (Internal Server Error) и вызываем функцию RenderApiError для отправки сообщения об ошибке клиенту.
Если метод json.Marshal возвращает ошибку при маршалинге найденной задачи, мы логируем эту ошибку, устанавливаем статус ответа 400 (Bad Request) и вызываем функцию RenderApiError для отправки сообщения об ошибке клиенту.
Если операция прошла успешно (без ошибок), мы устанавливаем заголовок ответа "Content-Type" на "application/json; charset=UTF-8", статус ответа 200 (OK) и записываем маршализованный объект JSON в тело ответа с помощью w.Write(resp).
Если при записи в тело ответа возникает ошибка, мы логируем её, устанавливаем статус ответа 400 (Bad Request) и вызываем функцию RenderApiError для отправки сообщения об ошибке клиенту.
*/
func (a *Api) GetTask(w http.ResponseWriter, r *http.Request, id int) {
	foundTask, err := a.repp.GetTask(id)
	log.Println("GetTask", "foundTask:", foundTask)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(foundTask)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(MarshallingError), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Println("error:", err)
		RenderApiError(w, fmt.Errorf(ResponseWriteError), http.StatusBadRequest)
		return
	}
}

// SigninHandler проверяет пароль и генерирует jwt token, если пароль верный
func (a *Api) SigninHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		RenderApiError(w, fmt.Errorf(ReadingError), http.StatusBadRequest)
		return
	}

	// берем пароль из request Body и записываем его в структуру signinRequest{} в поле Password
	reqBody := signinRequest{}
	err = json.Unmarshal(buf.Bytes(), &reqBody)
	if err != nil {
		RenderApiError(w, fmt.Errorf(UnMarshallingError), http.StatusBadRequest)
		return
	}

	secret := []byte(a.config.EncryptionSecretKey)
	hashedUserPassword := HashPassword([]byte(reqBody.Password), secret)
	hashedEnvPassword := HashPassword([]byte(a.config.AppPassword), secret)

	if hashedUserPassword != hashedEnvPassword {
		RenderApiError(w, fmt.Errorf("incorrect password"), http.StatusUnauthorized)
		return
	}

	// получаем подписанный токен
	tokenValue, err := createToken(reqBody.Password, a.config.EncryptionSecretKey)
	if err != nil {
		RenderApiError(w, fmt.Errorf(InternalServerError), http.StatusInternalServerError)
	}

	// записываем в response Body токен
	response := signinResponse{Token: tokenValue}
	respBody, err := json.Marshal(response)
	if err != nil {
		RenderApiError(w, fmt.Errorf(MarshallingError), http.StatusInternalServerError)
		return
	}

	WriteResponse(w, respBody)
	w.WriteHeader(http.StatusOK)
}

// добавим проверку аутентификации для API-запросов
func (a *Api) Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := a.config.AppPassword
		if len(pass) > 0 {
			var jwtFromRequest string
			// получаем Cookie
			cookie, err := r.Cookie("token")
			if err != nil {
				RenderApiError(w, fmt.Errorf("empty token"), http.StatusUnauthorized)
				return
			}
			jwtFromRequest = cookie.Value

			secret := []byte(a.config.EncryptionSecretKey)

			// валидация и проверка JWT-токена и парсим токен
			jwtToken, err := jwt.Parse(jwtFromRequest, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				RenderApiError(w, fmt.Errorf("invalid token"), http.StatusUnauthorized)
				return
			}

			// приводим поле Claims к типу jwt.MapClaims
			res, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				RenderApiError(w, fmt.Errorf("failed to typecast to jwt.MapCalims"), http.StatusUnauthorized)
				return
			}

			// ищем по ключу "password" т.к. Claims мапа
			pass := res["password"]
			_, ok = pass.(string)
			if !ok {
				RenderApiError(w, fmt.Errorf("failed to typecast to string"), http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
