package redisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(addr string) *Client {
	options := &redis.Options{
		Addr:                  addr,
		DialTimeout:           time.Second,
		ReadTimeout:           time.Second,
		WriteTimeout:          time.Second,
		ContextTimeoutEnabled: true,
	}

	redisClient := redis.NewClient(options)

	return &Client{
		client: redisClient,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}
