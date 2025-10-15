package recording

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Recorder manages call recordings
type Recorder struct {
	recordings map[string]*Recording
	mu         sync.RWMutex
	basePath   string
}

// Recording represents a call recording
type Recording struct {
	ID        string
	RoomID    string
	Filename  string
	StartedAt time.Time
	EndedAt   time.Time
	Active    bool
}

// NewRecorder creates a new Recorder instance
func NewRecorder(basePath string) *Recorder {
	// Create base path if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create recordings directory: %v", err))
	}
	
	return &Recorder{
		recordings: make(map[string]*Recording),
		basePath:   basePath,
	}
}

// StartRecording starts a new recording for a room
func (r *Recorder) StartRecording(roomID string) (*Recording, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Generate recording ID
	recordingID := uuid.New().String()
	
	// Generate filename
	filename := filepath.Join(r.basePath, fmt.Sprintf("%s_%s.webm", roomID, recordingID))
	
	// Create recording
	recording := &Recording{
		ID:        recordingID,
		RoomID:    roomID,
		Filename:  filename,
		StartedAt: time.Now(),
		Active:    true,
	}
	
	// Store recording
	r.recordings[recordingID] = recording
	
	// Create empty file
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create recording file: %v", err)
	}
	file.Close()
	
	return recording, nil
}

// StopRecording stops an active recording
func (r *Recorder) StopRecording(recordingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Find recording
	recording, exists := r.recordings[recordingID]
	if !exists {
		return fmt.Errorf("recording not found: %s", recordingID)
	}
	
	// Check if recording is active
	if !recording.Active {
		return fmt.Errorf("recording is not active: %s", recordingID)
	}
	
	// Update recording
	recording.Active = false
	recording.EndedAt = time.Now()
	
	return nil
}

// GetRecording returns a recording by ID
func (r *Recorder) GetRecording(recordingID string) (*Recording, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	recording, exists := r.recordings[recordingID]
	return recording, exists
}

// ListRecordings returns all recordings for a room
func (r *Recorder) ListRecordings(roomID string) []*Recording {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var recordings []*Recording
	for _, recording := range r.recordings {
		if recording.RoomID == roomID {
			// Return a copy to prevent external modification
			rec := *recording
			recordings = append(recordings, &rec)
		}
	}
	
	return recordings
}

// DeleteRecording deletes a recording file and removes it from the registry
func (r *Recorder) DeleteRecording(recordingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Find recording
	recording, exists := r.recordings[recordingID]
	if !exists {
		return fmt.Errorf("recording not found: %s", recordingID)
	}
	
	// Delete file
	if err := os.Remove(recording.Filename); err != nil {
		return fmt.Errorf("failed to delete recording file: %v", err)
	}
	
	// Remove from registry
	delete(r.recordings, recordingID)
	
	return nil
}

// GetRecordingFilePath returns the file path for a recording
func (r *Recorder) GetRecordingFilePath(recordingID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	recording, exists := r.recordings[recordingID]
	if !exists {
		return "", fmt.Errorf("recording not found: %s", recordingID)
	}
	
	return recording.Filename, nil
}