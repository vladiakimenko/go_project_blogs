package throttle

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var sharedClient *redis.Client

type RedisThrottler struct {
	redisClient *redis.Client
	Limit       int
	Window      time.Duration
	KeyPrefix   string
}

func InitRedis(addr, password string, db int) {
	sharedClient = redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		},
	)
}

func NewThrottler(keyPrefix string, limit int, window time.Duration) *RedisThrottler {
	return &RedisThrottler{
		redisClient: sharedClient,
		Limit:       limit,
		Window:      window,
		KeyPrefix:   keyPrefix,
	}
}

func (r *RedisThrottler) getKey(clientID string) string {
	return fmt.Sprintf("%s:%s", r.KeyPrefix, clientID)
}

func (r *RedisThrottler) Allow(ctx context.Context, clientID string) (bool, error) {
	key := r.getKey(clientID)
	now := time.Now().Unix()

	pipe := r.redisClient.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-int64(r.Window.Seconds())))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, r.Window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count, _ := countCmd.Result()
	return count < int64(r.Limit), nil
}

func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
