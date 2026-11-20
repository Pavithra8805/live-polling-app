package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var ErrDuplicateVoter = errors.New("voter has already voted")

type VoteResult struct {
	Counts map[string]int64
	Total  int64
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisStore) ApplyVote(ctx context.Context, pollID, optionID, voterID string) (VoteResult, error) {
	keys := []string{countsKey(pollID), votersKey(pollID), updatesChannel(pollID)}
	result, err := voteScript.Run(ctx, r.client, keys, voterID, optionID, pollID).Result()
	if err != nil {
		return VoteResult{}, err
	}
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return VoteResult{}, errors.New("unexpected vote script result")
	}
	accepted, ok := values[0].(int64)
	if !ok {
		return VoteResult{}, errors.New("unexpected vote script acceptance result")
	}
	if accepted == 0 {
		return VoteResult{}, ErrDuplicateVoter
	}
	payload, ok := values[1].(string)
	if !ok {
		return VoteResult{}, errors.New("unexpected vote script payload")
	}
	var response VoteResult
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		return VoteResult{}, fmt.Errorf("decode vote script payload: %w", err)
	}
	return response, nil
}

func (r *RedisStore) RollbackVote(ctx context.Context, pollID, optionID, voterID string) error {
	_, err := rollbackScript.Run(ctx, r.client,
		[]string{countsKey(pollID), votersKey(pollID), updatesChannel(pollID)},
		voterID, optionID,
	).Result()
	return err
}

func (r *RedisStore) GetCounts(ctx context.Context, pollID string) (map[string]int64, error) {
	values, err := r.client.HGetAll(ctx, countsKey(pollID)).Result()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(values))
	for optionID, value := range values {
		count, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Redis count for %s: %w", optionID, err)
		}
		counts[optionID] = count
	}
	return counts, nil
}

func (r *RedisStore) SetCounts(ctx context.Context, pollID string, counts map[string]int64) error {
	key := countsKey(pollID)
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, key)
	values := make(map[string]interface{}, len(counts))
	for optionID, count := range counts {
		values[optionID] = count
	}
	if len(values) > 0 {
		pipe.HSet(ctx, key, values)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) Subscribe(ctx context.Context) *redis.PubSub {
	return r.client.PSubscribe(ctx, "poll:*:updates")
}

func countsKey(pollID string) string {
	return "poll:" + pollID + ":counts"
}

func votersKey(pollID string) string {
	return "poll:" + pollID + ":voters"
}

func updatesChannel(pollID string) string {
	return "poll:" + pollID + ":updates"
}

var voteScript = redis.NewScript(`
local added = redis.call("SADD", KEYS[2], ARGV[1])
if added == 0 then
  return {0, "duplicate"}
end
redis.call("HINCRBY", KEYS[1], ARGV[2], 1)
local entries = redis.call("HGETALL", KEYS[1])
local counts = {}
local total = 0
for i = 1, #entries, 2 do
  counts[entries[i]] = tonumber(entries[i + 1])
  total = total + tonumber(entries[i + 1])
end
local payload = cjson.encode({counts = counts, total = total})
redis.call("PUBLISH", KEYS[3], payload)
return {1, payload}
`)

var rollbackScript = redis.NewScript(`
local removed = redis.call("SREM", KEYS[2], ARGV[1])
if removed == 0 then
  return 0
end
redis.call("HINCRBY", KEYS[1], ARGV[2], -1)
local entries = redis.call("HGETALL", KEYS[1])
local counts = {}
local total = 0
for i = 1, #entries, 2 do
  counts[entries[i]] = tonumber(entries[i + 1])
  total = total + tonumber(entries[i + 1])
end
local payload = cjson.encode({counts = counts, total = total})
redis.call("PUBLISH", KEYS[3], payload)
return 1
`)
