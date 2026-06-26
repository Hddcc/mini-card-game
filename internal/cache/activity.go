package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`

type RedisActivityCache struct {
	client *redis.Client
}

func NewRedisActivityCache(client *redis.Client) *RedisActivityCache {
	return &RedisActivityCache{client: client}
}

func (c *RedisActivityCache) Available() bool {
	return c != nil && c.client != nil
}

func (c *RedisActivityCache) AcquireLock(key string, token string, ttl time.Duration) (bool, error) {
	if !c.Available() {
		return true, nil
	}
	return c.client.SetNX(key, token, ttl).Result()
}

func (c *RedisActivityCache) ReleaseLock(key string, token string) error {
	if !c.Available() {
		return nil
	}
	return c.client.Eval(unlockScript, []string{key}, token).Err()
}

func ActivityPlayerDrawLockKey(playerID uint64) string {
	return fmt.Sprintf("activity:lottery:draw:player:%d", playerID)
}

func ActivityMessageScanLockKey() string {
	return "activity:lottery:job:message-scan"
}

func ActivityPrizePoolRefreshLockKey() string {
	return "activity:lottery:job:pool-refresh"
}

func ActivityConfigCacheKey(activityID uint64) string {
	return fmt.Sprintf("activity:lottery:config:%d", activityID)
}

func ActivityDailyPlayerCounterKey(activityID uint64, playerID uint64, day string) string {
	return fmt.Sprintf("activity:lottery:counter:player:%d:%d:%s", activityID, playerID, day)
}

func ActivityDailyIPCounterKey(activityID uint64, ip string, day string) string {
	return fmt.Sprintf("activity:lottery:counter:ip:%d:%s:%s", activityID, ip, day)
}

func ActivityBlacklistKey(targetType string, target string) string {
	return fmt.Sprintf("activity:lottery:blacklist:%s:%s", targetType, target)
}

func ActivityPrizePoolKey(activityID uint64) string {
	return fmt.Sprintf("activity:lottery:prize-pool:%d", activityID)
}

func (c *RedisActivityCache) GetJSON(key string, dest interface{}) (bool, error) {
	if !c.Available() {
		return false, nil
	}
	value, err := c.client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *RedisActivityCache) SetJSON(key string, value interface{}, ttl time.Duration) error {
	if !c.Available() {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(key, string(raw), ttl).Err()
}

func (c *RedisActivityCache) IncrWithTTL(key string, ttl time.Duration) (int64, error) {
	if !c.Available() {
		return 0, nil
	}
	count, err := c.client.Incr(key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		_ = c.client.Expire(key, ttl).Err()
	}
	return count, nil
}

func (c *RedisActivityCache) GetInt(key string) (int64, bool, error) {
	if !c.Available() {
		return 0, false, nil
	}
	value, err := c.client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil
		}
		return 0, false, err
	}
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return num, true, nil
}

func (c *RedisActivityCache) SetInt(key string, value int64, ttl time.Duration) error {
	if !c.Available() {
		return nil
	}
	return c.client.Set(key, strconv.FormatInt(value, 10), ttl).Err()
}

func (c *RedisActivityCache) HGetInt(key string, field string) (int64, bool, error) {
	if !c.Available() {
		return 0, false, nil
	}
	value, err := c.client.HGet(key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil
		}
		return 0, false, err
	}
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return num, true, nil
}

func (c *RedisActivityCache) HSetInt(key string, field string, value int64) error {
	if !c.Available() {
		return nil
	}
	return c.client.HSet(key, field, strconv.FormatInt(value, 10)).Err()
}

func (c *RedisActivityCache) HIncrBy(key string, field string, delta int64) (int64, error) {
	if !c.Available() {
		return 0, nil
	}
	return c.client.HIncrBy(key, field, delta).Result()
}
