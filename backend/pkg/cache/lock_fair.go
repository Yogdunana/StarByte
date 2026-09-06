package cache

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func waitTicketLeaseKey(queue, ticket string) string {
	return queue + ":" + ticket
}

func dropWaitTicket(ctx context.Context, rdb *redis.Client, queue, ticket string) error {
	if err := rdb.LRem(ctx, queue, 1, ticket).Err(); err != nil {
		return err
	}
	_ = rdb.Del(ctx, waitTicketLeaseKey(queue, ticket)).Err()
	return nil
}

func reclaimStaleHead(ctx context.Context, rdb *redis.Client, queue, head string) {
	if head == "" {
		return
	}
	n, err := rdb.Exists(ctx, waitTicketLeaseKey(queue, head)).Result()
	if err != nil || n > 0 {
		return
	}
	_ = reclaimHeadScript.Run(ctx, rdb, []string{queue}, head).Err()
}

// AcquireFair waits in a Redis list until it can take the lock (FIFO).
// Wait tickets carry a TTL lease so a dead waiter cannot wedge the queue.
func AcquireFair(ctx context.Context, rdb *redis.Client, name, owner string, ttl, wait time.Duration) (*Lock, error) {
	queue := lockPrefix + name + ":wait"
	ticket := uuid.NewString()
	if wait <= 0 {
		wait = 5 * time.Second
	}
	if err := rdb.Set(ctx, waitTicketLeaseKey(queue, ticket), "1", wait+time.Second).Err(); err != nil {
		return nil, err
	}
	if err := rdb.RPush(ctx, queue, ticket).Err(); err != nil {
		_ = rdb.Del(ctx, waitTicketLeaseKey(queue, ticket)).Err()
		return nil, err
	}
	qTTL := wait + ttl + time.Minute
	if qTTL < time.Minute {
		qTTL = time.Minute
	}
	_ = rdb.Expire(ctx, queue, qTTL).Err()
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		head, err := rdb.LIndex(ctx, queue, 0).Result()
		if err != nil && err != redis.Nil {
			_ = dropWaitTicket(ctx, rdb, queue, ticket)
			return nil, err
		}
		for head != "" && head != ticket {
			reclaimStaleHead(ctx, rdb, queue, head)
			next, nerr := rdb.LIndex(ctx, queue, 0).Result()
			if nerr != nil && nerr != redis.Nil {
				_ = dropWaitTicket(ctx, rdb, queue, ticket)
				return nil, nerr
			}
			if next == head {
				break
			}
			head = next
		}
		if head == ticket {
			lk, aerr := Acquire(ctx, rdb, name, owner, ttl)
			if aerr == nil {
				if err := dropWaitTicket(ctx, rdb, queue, ticket); err != nil {
					_ = lk.Unlock(ctx)
					return nil, err
				}
				return lk, nil
			}
			if aerr != ErrLockBusy {
				_ = dropWaitTicket(ctx, rdb, queue, ticket)
				return nil, aerr
			}
		}
		select {
		case <-ctx.Done():
			_ = dropWaitTicket(ctx, rdb, queue, ticket)
			return nil, ctx.Err()
		case <-time.After(40 * time.Millisecond):
		}
	}
	_ = dropWaitTicket(ctx, rdb, queue, ticket)
	return nil, ErrLockBusy
}
