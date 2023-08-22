package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"semaygateway.com/gatelogger"
	"semaygateway.com/gateparse"
)

// podman run -d -p 6379:6379 redis # for starting redus container using podman

type slidingWindow struct {
	redisClient *redis.Client
	keyPrefix   string
	rate        float64
	window      time.Duration
}

// var (
var ctx = context.Background()

func newSlidingWindow(redisClient *redis.Client, keyPrefix string, rate float64, window time.Duration) *slidingWindow {
	// Initialize  redis Clienet

	return &slidingWindow{
		redisClient: redisClient,
		keyPrefix:   keyPrefix,
		rate:        rate,
		window:      window,
	}
}

func (sw *slidingWindow) increment() error {
	now := time.Now().UnixNano()
	score := float64(now)
	member := fmt.Sprintf("%d", now)
	_, err := sw.redisClient.ZAdd(ctx, sw.keyPrefix, &redis.Z{
		Score:  score,
		Member: member,
	}).Result()
	if err != nil {
		return err
	}
	return nil
}

func (sw *slidingWindow) removeExpired() error {
	now := time.Now().UnixNano()
	minScore := float64(now) - sw.window.Seconds()*1e9
	_, err := sw.redisClient.ZRemRangeByScore(ctx, sw.keyPrefix, "0", fmt.Sprintf("%.0f", minScore)).Result()
	if err != nil {
		return err
	}
	return nil
}

func (sw *slidingWindow) countRequests() (int64, error) {
	now := time.Now().UnixNano()
	minScore := float64(now) - sw.window.Seconds()*1e9
	count, err := sw.redisClient.ZCount(ctx, sw.keyPrefix, fmt.Sprintf("%.0f", minScore), "+inf").Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

type SlidingWindowRateLimiter struct {
	slidingWindow *slidingWindow
}

func NewSlidingWindowRateLimiter(keyPrefix string) *SlidingWindowRateLimiter {

	//Gate Ratelimiting Configurations
	rateConfig, _ := gateparse.GetRateLimitConfig(keyPrefix)
	// get list of redis target
	redis_targets, _ := gateparse.GetRedisTargetLists()

	// Geting Redis Address
	redisAddress := redis_targets[(rateConfig.Redis - 1)]

	// Initalize the Redis Client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Setting the interval on with the rate limit is applied
	window := time.Duration(rateConfig.Interval * int(time.Second))

	// Setting the limit of the rate
	rate := float64(rateConfig.Limit)

	sw := newSlidingWindow(redisClient, keyPrefix, rate, window)
	return &SlidingWindowRateLimiter{
		slidingWindow: sw,
	}
}

func (rl *SlidingWindowRateLimiter) Allow() bool {
	// First Add to the counter
	err := rl.slidingWindow.increment()
	if err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return false
	}

	// Remove Expired  entries
	err = rl.slidingWindow.removeExpired()
	if err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return false
	}
	// Count the request
	count, err := rl.slidingWindow.countRequests()
	if err != nil {
		gatelogger.GateLoggerInfo(err.Error())
		return false
	}

	// fmt.Printf("number of request in the past 5 seconds is : %v \n", count)
	// set allowed requests requests within the provided interval
	allowedRequests := int64(rl.slidingWindow.rate)

	return count <= allowedRequests
}
