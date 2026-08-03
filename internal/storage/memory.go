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

func StoreToMemory(result monitor.Result) {
	checkedResults.mu.Lock()
	if key, ok := checkedResults.results[result.URL]; ok {
		key = append(key, result)
		checkedResults.results[result.URL] = key
	} else {
		checkedResults.results[result.URL] = []monitor.Result{result}
	}

	PrintResult(checkedResults.results)
	checkedResults.mu.Unlock()
}

func PrintResult(results map[string][]monitor.Result) {
	for key, arr := range results {
		fmt.Println(key)
		for _, val := range arr {
			fmt.Println(val)
		}
	}
}

func GetResults(url string) []monitor.Result {
	return checkedResults.results[url]
}

func GetAllResults() map[string][]monitor.Result {
	return checkedResults.results
}
func GetLastResult(url string) monitor.Result {
	return checkedResults.results[url][len(checkedResults.results[url])-1]
}
