package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"go_final_project/packages/config"
	"go_final_project/packages/db"
	"go_final_project/packages/handlers"
	"go_final_project/packages/repp"

	"github.com/go-chi/chi/v5"
	_ "modernc.org/sqlite"
)

var webDir = "./web/"

func main() {
	// создаем config, куда записываем пароль из переменной окружения и секретное слово
	config, err := config.NewConfig(
		os.Getenv("TODO_PASSWORD"),
		"my_very_secret_key",
		os.Getenv("TODO_PORT"),
	)
	if err != nil {
		log.Fatalf("Config error.")
	}
	db.DbConnection()

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		log.Println(err)
		return
	}
	tRepository := repp.NewTasksRepository(db)

	api := handlers.NewApi(&tRepository, config)

	r := chi.NewRouter()
	r.Handle("/*", http.FileServer(http.Dir(webDir)))
	r.Get("/api/nextdate", handlers.GetNextDay)

	// api, которое ожидается в этом задании
	// 1. POST /api/task создает таск
	// 2. GET /api/task возвращает ошибку (судя по тестам это желаемое поведение)
	// 3. GET /api/tasks возвращает набор тасков без фильтрации
	// 4. GET /api/tasks?search=... возвращает набор тасков с фильрацией по параметру search
	// 5. GET /api/tasks/{id} возвращает таск по id
	// r.HandleFunc("/api/task", api.TaskHandler) // get, post, put, delete
	r.HandleFunc("/api/task", api.Auth(api.TaskHandler))          // get, post, put, delete
	r.Get("/api/tasks", api.Auth(api.GetTasksHandler))            // search
	r.Get("/api/tasks/{id}", api.Auth(api.GetTaskByIdHandler))    // http://localhost:7540/api/tasks/257
	r.HandleFunc("/api/task/done", api.Auth(api.TaskDoneHandler)) // post И delete, здесь id - это параметр запроса

	r.Post("/api/signin", api.SigninHandler)

	r.Get("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Starting server on port: %s", config.ApiPort)
	if err := http.ListenAndServe(":"+config.ApiPort, r); err != nil {
		log.Printf("Start server error: %s", err.Error())
	}
}
