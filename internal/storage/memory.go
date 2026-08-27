package storage

import (
	"fmt"
	"sync"

	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
)

type MemoryStorage struct {
	mu      sync.RWMutex
	results map[string][]monitor.Result
}

var checkedResults = MemoryStorage{
	results: make(map[string][]monitor.Result),
}

// func StoreToMemory(result monitor.Result) {
// 	checkedResults.mu.Lock()
// 	if key, ok := checkedResults.results[result.ID]; ok {
// 		key = append(key, result)
// 		checkedResults.results[result.ID] = key
// 	} else {
// 		checkedResults.results[result.ID] = []monitor.Result{result}
// 	}

// 	PrintResult(checkedResults.results)
// 	checkedResults.mu.Unlock()
// }

func PrintResult(results map[string][]monitor.Result) {
	for key, arr := range results {
		fmt.Println(key)
		for _, val := range arr {
			fmt.Println(val)
		}
	}
}

// func GetResults(id string) []monitor.Result {
// 	fmt.Println("HELLO", checkedResults.results)

// 	return checkedResults.results[id]
// }

func GetAllResults() map[string][]monitor.Result {
	return checkedResults.results
}

// func GetLastResult(id string) monitor.Result {
// 	return checkedResults.results[id][len(checkedResults.results[id])-1]
// }
