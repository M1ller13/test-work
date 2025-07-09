package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndGetTask(t *testing.T) {
	tm := NewTaskManager()
	server := setupTestServer(tm)
	defer server.Close()

	resp, err := http.Post(server.URL+"/tasks", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		t.Fatal(err)
	}
	if task.ID == "" {
		t.Error("ожидался id задачи")
	}

	resp2, err := http.Get(server.URL + "/tasks/" + task.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	var got Task
	if err := json.NewDecoder(resp2.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != task.ID {
		t.Error("id задачи не совпадает")
	}
}

func TestDeleteTask(t *testing.T) {
	tm := NewTaskManager()
	server := setupTestServer(tm)
	defer server.Close()

	resp, _ := http.Post(server.URL+"/tasks", "application/json", nil)
	defer resp.Body.Close()
	var task Task
	json.NewDecoder(resp.Body).Decode(&task)

	req, _ := http.NewRequest("DELETE", server.URL+"/tasks/"+task.ID, nil)
	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNoContent {
		t.Errorf("ожидался статус 204, получено %d", resp2.StatusCode)
	}
}

func setupTestServer(tm *TaskManager) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			task := tm.CreateTask()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
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
	return httptest.NewServer(mux)
}
