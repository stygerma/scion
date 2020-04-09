// Copyright 2020 ETH Zurich
// Copyright 2020 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qosqueues

// import (
// 	"fmt"
// 	"sync"
// 	"testing"
// 	"time"
// )

// func TestBasic(t *testing.T) {
// 	bucket := TokenBucket{MaxBandWidth: 5 * 1024, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}

// 	fmt.Println(bucket)
// 	bucket.refill()
// 	fmt.Println(bucket)
// 	if(bucket.tokens != 0){
// 		t.Errorf("got %d, want %d", bucket.tokens, 0)
// 	}
// }

// func TestRefill(t *testing.T) {
// 	bucket := TokenBucket{MaxBandWidth: 5 * 1024, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}

// 	fmt.Println(bucket)
// 	time.Sleep(time.Millisecond * 2000)
// 	bucket.refill()
// 	fmt.Println(bucket)
// 	if(bucket.tokens != (5 * 1024) * 2){
// 		t.Errorf("got %d, want %d", bucket.tokens, (5 * 1024) * 2)
// 	}
// }

// func TestRefillTwice(t *testing.T) {
// 	bucket := TokenBucket{MaxBandWidth: 5 * 1024, tokens: 0, lastRefill: time.Now(), mutex: &sync.Mutex{}}

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
