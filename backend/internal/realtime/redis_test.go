package realtime

import (
	"context"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *RedisStore) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { client.Close() })
	return server, NewRedisStore(client)
}

func TestApplyVoteRejectsDuplicateVoter(t *testing.T) {
	_, store := newTestRedis(t)
	ctx := context.Background()

	first, err := store.ApplyVote(ctx, "poll-1", "option-1", "voter-1")
	if err != nil {
		t.Fatalf("first vote error = %v", err)
	}
	if first.Total != 1 || first.Counts["option-1"] != 1 {
		t.Fatalf("first vote result = %+v", first)
	}

	if _, err := store.ApplyVote(ctx, "poll-1", "option-2", "voter-1"); err != ErrDuplicateVoter {
		t.Fatalf("duplicate vote error = %v, want ErrDuplicateVoter", err)
	}
}

func TestConcurrentVotesHaveExactCounts(t *testing.T) {
	_, store := newTestRedis(t)
	ctx := context.Background()
	const voters = 200

	var wg sync.WaitGroup
	errs := make(chan error, voters)
	for index := 0; index < voters; index++ {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			option := "option-1"
			if index%2 == 1 {
				option = "option-2"
			}
			_, err := store.ApplyVote(ctx, "poll-concurrent", option, "voter-"+itoa(index))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent vote error = %v", err)
		}
	}

	counts, err := store.GetCounts(ctx, "poll-concurrent")
	if err != nil {
		t.Fatalf("GetCounts() error = %v", err)
	}
	if counts["option-1"] != 100 || counts["option-2"] != 100 {
		t.Fatalf("counts = %+v", counts)
	}

	var duplicateWG sync.WaitGroup
	duplicateErrors := make(chan error, voters)
	for index := 0; index < voters; index++ {
		index := index
		duplicateWG.Add(1)
		go func() {
			defer duplicateWG.Done()
			option := "option-1"
			if index%2 == 1 {
				option = "option-2"
			}
			_, err := store.ApplyVote(ctx, "poll-concurrent", option, "voter-"+itoa(index))
			duplicateErrors <- err
		}()
	}
	duplicateWG.Wait()
	close(duplicateErrors)
	for err := range duplicateErrors {
		if err != ErrDuplicateVoter {
			t.Fatalf("duplicate concurrency error = %v", err)
		}
	}
}

func itoa(value int) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = digits[value%10]
		value /= 10
	}
	return string(buffer[index:])
}
