package db

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type VideoJob struct {
	ID         string    `json:"id"`
	TemplateID string    `json:"template_id"`
	Status     string    `json:"status"` // "rendering", "done", "error"
	FilePath   string    `json:"file_path"`
	Category   string    `json:"category"`
	Archived   bool      `json:"archived"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Error      string    `json:"error,omitempty"`
}

var (
	dbMutex sync.Mutex
	dbFile  = "data/videos_db.json"
)

func InitDB() error {
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return saveDB(map[string]VideoJob{})
	}
	return nil
}

func ReadDB() (map[string]VideoJob, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	data, err := os.ReadFile(dbFile)
	if err != nil {
		return nil, err
	}

	var jobs map[string]VideoJob
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func saveDB(jobs map[string]VideoJob) error {
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dbFile, data, 0644)
}

func AddVideoJob(job VideoJob) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	data, _ := os.ReadFile(dbFile)
	var jobs map[string]VideoJob
	if err := json.Unmarshal(data, &jobs); err != nil {
		jobs = make(map[string]VideoJob)
	}

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	jobs[job.ID] = job

	out, _ := json.MarshalIndent(jobs, "", "  ")
	return os.WriteFile(dbFile, out, 0644)
}

func UpdateVideoJobStatus(id, status, errorMsg string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	data, _ := os.ReadFile(dbFile)
	var jobs map[string]VideoJob
	json.Unmarshal(data, &jobs)

	if job, ok := jobs[id]; ok {
		job.Status = status
		if errorMsg != "" {
			job.Error = errorMsg
		}
		job.UpdatedAt = time.Now()
		jobs[id] = job
		out, _ := json.MarshalIndent(jobs, "", "  ")
		return os.WriteFile(dbFile, out, 0644)
	}
	return fmt.Errorf("job not found")
}

func UpdateVideoJobMeta(id, category string, archived bool) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	data, _ := os.ReadFile(dbFile)
	var jobs map[string]VideoJob
	json.Unmarshal(data, &jobs)

	if job, ok := jobs[id]; ok {
		job.Category = category
		job.Archived = archived
		job.UpdatedAt = time.Now()
		jobs[id] = job
		out, _ := json.MarshalIndent(jobs, "", "  ")
		return os.WriteFile(dbFile, out, 0644)
	}
	return fmt.Errorf("job not found")
}

func DeleteVideoJob(id string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	data, _ := os.ReadFile(dbFile)
	var jobs map[string]VideoJob
	json.Unmarshal(data, &jobs)

	if job, ok := jobs[id]; ok {
		// Attempt to delete the file
		if job.FilePath != "" {
			os.Remove(job.FilePath)
		}
		delete(jobs, id)
		out, _ := json.MarshalIndent(jobs, "", "  ")
		return os.WriteFile(dbFile, out, 0644)
	}
	return nil
}
