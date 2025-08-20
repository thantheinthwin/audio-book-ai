package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	APIBaseURL = "http://localhost:8080"
	AuthToken  = "your-auth-token-here" // Replace with actual token
)

// TestUploadFlow demonstrates the complete audiobook creation flow
func main() {
	log.Println("Starting audiobook creation flow test...")

	// Step 1: Create upload session
	uploadID, err := createUploadSession()
	if err != nil {
		log.Fatalf("Failed to create upload session: %v", err)
	}
	log.Printf("Created upload session: %s", uploadID)

	// Step 2: Upload audio file
	fileID, err := uploadAudioFile(uploadID, "test-audio.mp3")
	if err != nil {
		log.Fatalf("Failed to upload audio file: %v", err)
	}
	log.Printf("Uploaded audio file: %s", fileID)

	// Step 3: Create audiobook from upload
	audiobookID, err := createAudioBook(uploadID)
	if err != nil {
		log.Fatalf("Failed to create audiobook: %v", err)
	}
	log.Printf("Created audiobook: %s", audiobookID)

	// Step 4: Monitor job progress
	monitorJobProgress(audiobookID)
}

// createUploadSession creates a new upload session
func createUploadSession() (string, error) {
	payload := map[string]interface{}{
		"upload_type":      "single",
		"total_files":      1,
		"total_size_bytes": 1024000, // 1MB
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(APIBaseURL+"/api/v1/admin/uploads", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create upload session: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["upload_id"].(string), nil
}

// uploadAudioFile uploads an audio file to the upload session
func uploadAudioFile(uploadID, filename string) (string, error) {
	// Create a test file if it doesn't exist
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		createTestAudioFile(filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	writer.Close()

	url := fmt.Sprintf("%s/api/v1/admin/uploads/%s/files", APIBaseURL, uploadID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+AuthToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload file: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["file_id"].(string), nil
}

// createAudioBook creates an audiobook from the completed upload
func createAudioBook(uploadID string) (string, error) {
	payload := map[string]interface{}{
		"upload_id":   uploadID,
		"title":       "Test Audio Book",
		"author":      "Test Author",
		"description": "A test audio book for demonstration",
		"language":    "en",
		"is_public":   false,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(APIBaseURL+"/api/v1/admin/audiobooks", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create audiobook: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["audiobook_id"].(string), nil
}

// monitorJobProgress monitors the progress of processing jobs
func monitorJobProgress(audiobookID string) {
	log.Println("Monitoring job progress...")

	for i := 0; i < 30; i++ { // Monitor for 5 minutes
		time.Sleep(10 * time.Second)

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/admin/audiobooks/%s/jobs", APIBaseURL, audiobookID))
		if err != nil {
			log.Printf("Failed to get job status: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to get job status: %d", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		overallStatus := result["overall_status"].(string)
		progress := result["progress"].(float64)
		completedJobs := int(result["completed_jobs"].(float64))
		totalJobs := int(result["total_jobs"].(float64))

		log.Printf("Status: %s, Progress: %.2f%% (%d/%d jobs completed)",
			overallStatus, progress*100, completedJobs, totalJobs)

		if overallStatus == "completed" {
			log.Println("All jobs completed successfully!")
			break
		} else if overallStatus == "failed" {
			log.Println("Some jobs failed!")
			break
		}
	}
}

// createTestAudioFile creates a dummy audio file for testing
func createTestAudioFile(filename string) {
	// Create a simple MP3 header (this is just for testing)
	header := []byte{
		0xFF, 0xFB, 0x90, 0x44, // MP3 sync word
	}

	// Add some dummy data
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	file.Write(header)
	file.Write(data)

	log.Printf("Created test audio file: %s", filename)
}

// Helper function to add auth headers
func addAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+AuthToken)
}
