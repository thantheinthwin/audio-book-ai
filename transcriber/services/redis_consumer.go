package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisConsumer handles consuming transcription jobs from Redis queues
type RedisConsumer struct {
	client *redis.Client
	prefix string
	config *Config
}

// JobMessage represents a transcription job message from the queue
type JobMessage struct {
	ID          uuid.UUID              `json:"id"`
	AudiobookID uuid.UUID              `json:"audiobook_id"`
	JobType     string                 `json:"job_type"`
	FilePath    *string                `json:"file_path,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Priority    int                    `json:"priority"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewRedisConsumer creates a new Redis consumer
func NewRedisConsumer(redisURL, prefix string, config *Config) (*RedisConsumer, error) {
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

	return &RedisConsumer{
		client: client,
		prefix: prefix,
		config: config,
	}, nil
}

// Close closes the Redis connection
func (r *RedisConsumer) Close() error {
	return r.client.Close()
}

// getQueueName returns the full queue name with prefix
func (r *RedisConsumer) getQueueName(jobType string) string {
	return fmt.Sprintf("%s:queue:%s", r.prefix, jobType)
}

// getProcessingQueueName returns the processing queue name
func (r *RedisConsumer) getProcessingQueueName(jobType string) string {
	return fmt.Sprintf("%s:processing:%s", r.prefix, jobType)
}

// getFailedQueueName returns the failed queue name
func (r *RedisConsumer) getFailedQueueName(jobType string) string {
	return fmt.Sprintf("%s:failed:%s", r.prefix, jobType)
}

// ConsumeJobs starts consuming transcription jobs from the queue
func (r *RedisConsumer) ConsumeJobs(ctx context.Context, jobType string, processor func(JobMessage) error) error {
	queueName := r.getQueueName(jobType)
	processingQueueName := r.getProcessingQueueName(jobType)
	failedQueueName := r.getFailedQueueName(jobType)

	log.Printf("Starting transcription consumer for queue: %s", queueName)

	for {
		// Check for context cancellation before each iteration
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping consumer: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		// Create a shorter timeout for the blocking operation to allow for context cancellation
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		result, err := r.client.BZPopMin(timeoutCtx, 2*time.Second, queueName).Result()
		cancel()

		if err != nil {
			if err == redis.Nil {
				// No jobs available, continue
				continue
			}
			if ctx.Err() != nil {
				// Context was cancelled during the operation
				log.Printf("Context cancelled during Redis operation: %v", ctx.Err())
				return ctx.Err()
			}
			log.Printf("Error getting job from queue: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if result == nil {
			continue
		}

		// Parse job message
		var message JobMessage
		memberStr, ok := result.Member.(string)
		if !ok {
			log.Printf("Invalid member type in queue")
			continue
		}
		if err := json.Unmarshal([]byte(memberStr), &message); err != nil {
			log.Printf("Error unmarshaling job message: %v", err)
			continue
		}

		log.Printf("Processing transcription job %s for audiobook %s", message.ID, message.AudiobookID)

		// Move to processing queue
		processingBytes, _ := json.Marshal(message)
		r.client.ZAdd(ctx, processingQueueName, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: processingBytes,
		})

		// Process the job
		if err := processor(message); err != nil {
			log.Printf("Error processing transcription job %s: %v", message.ID, err)

			// Handle retry logic
			if message.RetryCount < message.MaxRetries {
				message.RetryCount++
				message.Priority = 10 // Higher priority for retries

				// Add back to main queue with delay
				retryBytes, _ := json.Marshal(message)
				score := float64(time.Now().Add(time.Duration(message.RetryCount*30) * time.Second).Unix())
				r.client.ZAdd(ctx, queueName, redis.Z{
					Score:  score,
					Member: retryBytes,
				})
			} else {
				// Move to failed queue
				failedBytes, _ := json.Marshal(message)
				r.client.ZAdd(ctx, failedQueueName, redis.Z{
					Score:  float64(time.Now().Unix()),
					Member: failedBytes,
				})
			}
		}

		// Remove from processing queue
		r.client.ZRem(ctx, processingQueueName, processingBytes)
	}
}

// GetQueueStats returns statistics about the transcription queues
func (r *RedisConsumer) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	jobTypes := []string{"transcribe"}

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

		stats[jobType] = map[string]interface{}{
			"pending":    pending,
			"processing": processing,
			"failed":     failed,
		}
	}

	return stats, nil
}
