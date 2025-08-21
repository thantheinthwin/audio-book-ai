package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"audio-book-ai/api/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisQueueService handles Redis queue operations
type RedisQueueService struct {
	client *redis.Client
	prefix string
}

// JobMessage represents a job message in the queue
type JobMessage struct {
	ID          uuid.UUID              `json:"id"`
	AudiobookID uuid.UUID              `json:"audiobook_id"`
	JobType     models.JobType         `json:"job_type"`
	FilePath    *string                `json:"file_path,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Priority    int                    `json:"priority"` // Higher number = higher priority
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewRedisQueueService creates a new Redis queue service
func NewRedisQueueService(redisURL, prefix string) (*RedisQueueService, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisQueueService{
		client: client,
		prefix: prefix,
	}, nil
}

// Close closes the Redis connection
func (r *RedisQueueService) Close() error {
	return r.client.Close()
}

// getQueueName returns the full queue name with prefix
func (r *RedisQueueService) getQueueName(jobType models.JobType) string {
	return fmt.Sprintf("%s:queue:%s", r.prefix, jobType)
}

// getProcessingQueueName returns the processing queue name
func (r *RedisQueueService) getProcessingQueueName(jobType models.JobType) string {
	return fmt.Sprintf("%s:processing:%s", r.prefix, jobType)
}

// getFailedQueueName returns the failed queue name
func (r *RedisQueueService) getFailedQueueName(jobType models.JobType) string {
	return fmt.Sprintf("%s:failed:%s", r.prefix, jobType)
}

// EnqueueJob adds a job to the appropriate Redis queue
func (r *RedisQueueService) EnqueueJob(ctx context.Context, job *models.ProcessingJob, filePath *string) error {
	queueName := r.getQueueName(job.JobType)

	message := JobMessage{
		ID:          job.ID,
		AudiobookID: job.AudiobookID,
		JobType:     job.JobType,
		FilePath:    filePath,
		CreatedAt:   job.CreatedAt,
		Priority:    1, // Default priority
		RetryCount:  0,
		MaxRetries:  3,
		Metadata: map[string]interface{}{
			"redis_job_id": job.RedisJobID,
		},
	}

	// Serialize message
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal job message: %v", err)
	}

	// Add to queue with priority (using sorted set)
	score := float64(time.Now().Unix()) + float64(message.Priority)
	err = r.client.ZAdd(ctx, queueName, redis.Z{
		Score:  score,
		Member: messageBytes,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue job: %v", err)
	}

	return nil
}

// EnqueueTranscriptionJob adds a transcription job to the queue
func (r *RedisQueueService) EnqueueTranscriptionJob(ctx context.Context, job *models.ProcessingJob, filePath string) error {
	return r.EnqueueJob(ctx, job, &filePath)
}

// EnqueueAIJob adds an AI processing job to the queue
func (r *RedisQueueService) EnqueueAIJob(ctx context.Context, job *models.ProcessingJob) error {
	return r.EnqueueJob(ctx, job, nil)
}

// GetQueueStats returns statistics about the queues
func (r *RedisQueueService) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	jobTypes := []models.JobType{
		models.JobTypeTranscribe,
		models.JobTypeSummarize,
		models.JobTypeEmbed,
	}

	stats := make(map[string]interface{})

	for _, jobType := range jobTypes {
		queueName := r.getQueueName(jobType)
		processingQueueName := r.getProcessingQueueName(jobType)
		failedQueueName := r.getFailedQueueName(jobType)

		// Get queue sizes
		pending, err := r.client.ZCard(ctx, queueName).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get pending queue size: %v", err)
		}

		processing, err := r.client.ZCard(ctx, processingQueueName).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get processing queue size: %v", err)
		}

		failed, err := r.client.ZCard(ctx, failedQueueName).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get failed queue size: %v", err)
		}

		stats[string(jobType)] = map[string]interface{}{
			"pending":    pending,
			"processing": processing,
			"failed":     failed,
		}
	}

	return stats, nil
}

// ClearQueue clears all jobs from a specific queue
func (r *RedisQueueService) ClearQueue(ctx context.Context, jobType models.JobType) error {
	queueName := r.getQueueName(jobType)
	return r.client.Del(ctx, queueName).Err()
}

// RetryFailedJob moves a job from failed queue back to main queue
func (r *RedisQueueService) RetryFailedJob(ctx context.Context, jobType models.JobType, jobID uuid.UUID) error {
	failedQueueName := r.getFailedQueueName(jobType)
	queueName := r.getQueueName(jobType)

	// Get the failed job
	failedJobs, err := r.client.ZRange(ctx, failedQueueName, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get failed jobs: %v", err)
	}

	for _, jobBytes := range failedJobs {
		var message JobMessage
		if err := json.Unmarshal([]byte(jobBytes), &message); err != nil {
			continue
		}

		if message.ID == jobID {
			// Remove from failed queue
			r.client.ZRem(ctx, failedQueueName, jobBytes)

			// Increment retry count
			message.RetryCount++

			// Add back to main queue with higher priority
			message.Priority = 10
			newMessageBytes, _ := json.Marshal(message)
			score := float64(time.Now().Unix()) + float64(message.Priority)

			return r.client.ZAdd(ctx, queueName, redis.Z{
				Score:  score,
				Member: newMessageBytes,
			}).Err()
		}
	}

	return fmt.Errorf("job not found in failed queue")
}
