package main

// import (
// 	"fmt"
// 	"sync"
// 	"testing"
// 	"time"
// )

// func TestBasic(t *testing.T) {
// 	bucket := tokenBucket{MaxBandWidth: 5 * 1024, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}

// 	fmt.Println(bucket)
// 	bucket.refill()
// 	fmt.Println(bucket)
// 	if(bucket.tokens != 0){
// 		t.Errorf("got %d, want %d", bucket.tokens, 0)
// 	}
// }

// func TestRefill(t *testing.T) {
// 	bucket := tokenBucket{MaxBandWidth: 5 * 1024, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}

// 	fmt.Println(bucket)
// 	time.Sleep(time.Millisecond * 2000)
// 	bucket.refill()
// 	fmt.Println(bucket)
// 	if(bucket.tokens != (5 * 1024) * 2){
// 		t.Errorf("got %d, want %d", bucket.tokens, (5 * 1024) * 2)
// 	}
// }

// func TestRefillTwice(t *testing.T) {
// 	bucket := tokenBucket{MaxBandWidth: 5 * 1024, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}

// 	fmt.Println(bucket)
// 	time.Sleep(time.Millisecond * 2000)
// 	bucket.refill()
// 	time.Sleep(time.Millisecond * 2000)
// 	bucket.refill()
// 	fmt.Println(bucket)
// 	if(bucket.tokens != (5 * 1024) / 10 * 2 * 2){
// 		t.Errorf("got %d, want %d", bucket.tokens, (5 * 1024) / 10 * 2 * 2)
// 	}
// }