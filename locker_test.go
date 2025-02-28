package lockit

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"
)

func TestTryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Del(context.Background(), "locker")
	locker := NewRedisLocker(rdb)
	var f func(*sync.WaitGroup, string)
	f = func(wg *sync.WaitGroup, name string) {
		defer wg.Done()
		ctx := context.Background()
		for {
			lock, err := locker.TryLock(ctx, "locker", name, 1*time.Minute)
			if err != nil {
				panic(err)
			}
			if lock {
				fmt.Println(name + " get the locker!")
				time.Sleep(1 * time.Second)
				err = locker.Unlock(ctx, "locker", name)
				if err != nil {
					panic(err)
				}
				break
			} else {
				fmt.Println(name + " miss the locker🥲")
				time.Sleep(5 * time.Second)
			}
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go f(wg, "proc1")
	go f(wg, "proc2")
	wg.Wait()
}
