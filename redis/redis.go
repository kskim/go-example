package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	err = client.Set("mykey", "value", time.Second*10).Err()
	if err != nil {
		panic(err)
	}

	result, err := client.Get("mykey").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	//랭킹
	client.ZAdd("myrank", &redis.Z{
		Score:  100,
		Member: "a",
	})

	client.ZAdd("myrank", &redis.Z{
		Score:  200,
		Member: "b",
	})

	// 0~1 위까지
	strings, err := client.ZRange("myrank", 0, 1).Result()
	fmt.Println(strings)

	//오름차순
	rank, err := client.ZRank("myrank", "a").Result()
	fmt.Println(rank)

	//내림차순
	zscore, err := client.ZRevRank("myrank", "b").Result()
	fmt.Println(zscore)
}
