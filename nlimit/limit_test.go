package nlimit

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestLimiter_Memory(t *testing.T) {
	// 允许 2 次/秒
	l, err := New("2-S")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ctx := context.Background()
	key := "test_user"

	// 第 1 次: 允许
	c, err := l.Allow(ctx, key)
	if err != nil || c.Reached {
		t.Errorf("1st request should be allowed")
	}

	// 第 2 次: 允许
	c, err = l.Allow(ctx, key)
	if err != nil || c.Reached {
		t.Errorf("2nd request should be allowed")
	}

	// 第 3 次: 拒绝
	c, err = l.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow failed: %v", err)
	}
	if !c.Reached {
		t.Errorf("3rd request should be blocked")
	}
}

func TestLimiter_Redis(t *testing.T) {
	// 连接本地 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping Redis test: %v", err)
	}

	// 清理旧数据
	rdb.FlushDB(ctx)

	// 允许 5 次/分钟
	l, err := New("5-M", WithRedis(rdb), WithPrefix("test_limit"))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	key := "test_redis_user"

	// 消耗完 5 次
	for i := 0; i < 5; i++ {
		c, _ := l.Allow(ctx, key)
		if c.Reached {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 第 6 次: 拒绝
	c, _ := l.Allow(ctx, key)
	if !c.Reached {
		t.Errorf("6th request should be blocked")
	}
}

func TestLimiter_Concurrent(t *testing.T) {
	// 并发测试 (Memory)
	l, _ := New("1000-S")
	ctx := context.Background()
	key := "concurrent_user"

	// 启动 10 个 goroutine，每个请求 10 次
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				l.Allow(ctx, key)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 检查最终计数
	c, _ := l.Allow(ctx, key)
	// 之前跑了 100 次，这次是第 101 次
	// Limit 是 1000，所以应该还没 Reached
	if c.Reached {
		t.Errorf("Should not be reached yet")
	}
}
