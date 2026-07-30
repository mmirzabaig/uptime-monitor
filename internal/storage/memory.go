package storage

import (
	"fmt"

	"github.com/mmirzabaig/uptime-monitor/internal/monitor"
)

var checkedResults = make(map[string][]monitor.Result)

func StoreToMemory(result monitor.Result) {
	if key, ok := checkedResults[result.URL]; ok {
		key = append(key, result)
		checkedResults[result.URL] = key
	} else {
		checkedResults[result.URL] = []monitor.Result{result}
	}
}

func PrintResult(results map[string][]monitor.Result) {
	for key, arr := range results {
		fmt.Println(key)
		for _, val := range arr {
			fmt.Println(val)
		}
	}
}
