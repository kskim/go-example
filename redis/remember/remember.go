package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"strconv"
	"time"
)

func main() {
	testKeyName := "tx"

	//redis 접속
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	//key 삭제
	client.Del(testKeyName)

	// go 루틴으로 멀티 스레드 동시 접근 환경 테스트, 단일 키에 대한 동시 접근시 안전하게 캐시를 생성하고 가져올수 있는지 테스트 한다.
	// 가장 먼저 설정한 값을 사용하고 뒤에 실행되는 set 명령이 실행 되지 않아야 한다.
	for i := 0; i < 100; i++ {
		name := strconv.Itoa(i)
		fmt.Printf("run thread %s, set value try %s  \n", name, name)
		go remember(client, testKeyName, func() string {
			return name
		}, 0)
	}

	//몇초 이후 마지막에 설정된 값을 표시한다.
	time.Sleep(time.Second * 3)
	result, err := client.Get(testKeyName).Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("finally cached data : %s \n", result)
}

// 캐시 값을 가져오거나 없는경우 callback 함수를 실행해서 값을 설정하여 가져온다.
// get value already and set value from callback if cache value is null
func remember(client *redis.Client, key string, callback func() string, expiration time.Duration) string {
	//캐시에 값이 있는지 가져온다.
	//키가 없는경우
	if callback != nil {
		//callback 이 설정된경우 실행하여 캐시 할 값을 가져온다.
		var cachedValue string

		//watch 명령 실행. 해당 key 의 변경 상황을 감지한다.
		//key 에 값을 추가한다. 여기서 transaction 이 발생하기때문에 동시성 접근 문제를 해결할수 있다.
		err := client.Watch(func(tx *redis.Tx) error {
			alreadyValue, err := tx.Get(key).Result()
			if err == redis.Nil {
				cachedValue = callback()
				fmt.Printf("redis set is : %s \n", cachedValue)
				_, err := tx.Set(key, cachedValue, expiration).Result()
				return err
			} else {
				cachedValue = alreadyValue
			}
			return err
		}, key)

		if err == nil {
			//tx 를 얻지 못한 다른 쓰레드에서는 오류가 발생한다. 오류가 발생하지 않으면 빠져나간다.
			return cachedValue
		} else {
			panic(err)
		}
	}
	return ""
}
