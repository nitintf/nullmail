package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
	ctx    context.Context
}

type Config struct {
	Addr     string
	Password string
	DB       int
}

func NewClient(config Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &Client{
		client: rdb,
		ctx:    context.Background(),
	}
}

func NewClientFromEnv() *Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		password = "dev123"
	}

	return NewClient(Config{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
}

func (c *Client) Ping() error {
	_, err := c.client.Ping(c.ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	slog.Info("Redis connection successful")
	return nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) StoreEmail(emailID string, emailData interface{}) error {
	data, err := json.Marshal(emailData)
	if err != nil {
		return fmt.Errorf("failed to marshal email data: %w", err)
	}

	key := fmt.Sprintf("nullmail:email:%s", emailID)
	err = c.client.Set(c.ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store email in redis: %w", err)
	}

	// Add to email list for easy retrieval
	err = c.client.LPush(c.ctx, "nullmail:emails", emailID).Err()
	if err != nil {
		return fmt.Errorf("failed to add email to list: %w", err)
	}

	slog.Info("Email stored in Redis", "id", emailID, "key", key)
	return nil
}

func (c *Client) StoreEmailWithRecipients(emailID string, emailData interface{}, recipients []string) error {
	// First store the email using the existing method
	err := c.StoreEmail(emailID, emailData)
	if err != nil {
		return err
	}

	// Index by recipients for efficient lookup
	for _, recipient := range recipients {
		recipientKey := fmt.Sprintf("emails:%s", recipient)
		err = c.client.LPush(c.ctx, recipientKey, emailID).Err()
		if err != nil {
			slog.Error("Failed to add email to recipient index", "recipient", recipient, "emailID", emailID, "error", err)
			continue
		}

		err = c.client.Expire(c.ctx, recipientKey, 24*time.Hour).Err()
		if err != nil {
			slog.Warn("Failed to set TTL for recipient list", "recipient", recipient, "error", err)
		}

		slog.Debug("Email indexed for recipient", "recipient", recipient, "emailID", emailID)
	}

	slog.Info("Email stored with recipient indexing", "id", emailID, "recipients", recipients)
	return nil
}

func (c *Client) GetEmail(emailID string) (string, error) {
	key := fmt.Sprintf("nullmail:email:%s", emailID)
	result, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("email not found: %s", emailID)
	} else if err != nil {
		return "", fmt.Errorf("failed to get email from redis: %w", err)
	}
	return result, nil
}

func (c *Client) GetAllEmails() ([]string, error) {
	result, err := c.client.LRange(c.ctx, "nullmail:emails", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get email list: %w", err)
	}
	return result, nil
}

// GetEmailsForRecipient retrieves all email IDs for a specific recipient
func (c *Client) GetEmailsForRecipient(recipient string) ([]string, error) {
	recipientKey := fmt.Sprintf("emails:%s", recipient)
	result, err := c.client.LRange(c.ctx, recipientKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get emails for recipient %s: %w", recipient, err)
	}
	return result, nil
}

func (c *Client) QueueEmail(queueName string, emailData interface{}) error {
	data, err := json.Marshal(emailData)
	if err != nil {
		return fmt.Errorf("failed to marshal email for queue: %w", err)
	}

	key := fmt.Sprintf("nullmail:queue:%s", queueName)
	err = c.client.LPush(c.ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}

	slog.Debug("Email queued", "queue", queueName, "key", key)
	return nil
}

func (c *Client) DequeueEmail(queueName string) (string, error) {
	key := fmt.Sprintf("nullmail:queue:%s", queueName)
	result, err := c.client.RPop(c.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // No items in queue
	} else if err != nil {
		return "", fmt.Errorf("failed to dequeue email: %w", err)
	}
	return result, nil
}

func (c *Client) IncrementEmailCount(countType string) error {
	key := fmt.Sprintf("nullmail:stats:%s", countType)
	err := c.client.Incr(c.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to increment counter: %w", err)
	}
	return nil
}

func (c *Client) GetEmailCount(countType string) (int64, error) {
	key := fmt.Sprintf("nullmail:stats:%s", countType)
	count, err := c.client.Get(c.ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}
	return count, nil
}
