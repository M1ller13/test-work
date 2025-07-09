package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Статусы задачи
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// Описание задачи
// Можно добавить новые поля при необходимости
type Task struct {
	ID        string      `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	Duration  float64     `json:"duration"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// Хранилище задач
type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
	}
}

// Создать новую задачу
func (tm *TaskManager) CreateTask() *Task {
	id := strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.Itoa(rand.Intn(10000))
	task := &Task{
		ID:        id,
		CreatedAt: time.Now(),
		Status:    StatusPending,
	}
	tm.mu.Lock()
	tm.tasks[id] = task
	tm.mu.Unlock()
	go tm.runTask(task)
	return task
}

// Имитация долгой работы
func (tm *TaskManager) runTask(task *Task) {
	tm.updateStatus(task.ID, StatusRunning)
	dur := time.Duration(180+rand.Intn(120)) * time.Second
	start := time.Now()
	time.Sleep(dur)
	task.Duration = time.Since(start).Seconds()
	task.Result = "Задача выполнена"
	tm.updateStatus(task.ID, StatusCompleted)
}

func (tm *TaskManager) updateStatus(id, status string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if task, ok := tm.tasks[id]; ok {
		task.Status = status
	}
}

// Получить задачу по id
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[id]
	return task, ok
}

// Удалить задачу
func (tm *TaskManager) DeleteTask(id string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if _, ok := tm.tasks[id]; ok {
		delete(tm.tasks, id)
		return true
	}
	return false
}

func main() {
	tm := NewTaskManager()

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			task := tm.CreateTask()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/tasks/"):]
		if id == "" {
			http.Error(w, "Нужен id задачи", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			task, ok := tm.GetTask(id)
			if !ok {
				http.Error(w, "Задача не найдена", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
		case http.MethodDelete:
			if tm.DeleteTask(id) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.Error(w, "Задача не найдена", http.StatusNotFound)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
