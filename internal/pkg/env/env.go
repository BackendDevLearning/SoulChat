package env

import (
	"fmt"
	"os"
)

// 环境类型枚举
type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

var currentEnv Environment

func init() {
	// 从环境变量读取，默认开发环境
	envStr := os.Getenv("ChatAPP_ENV")
	switch envStr {
	case "production":
		currentEnv = Production
	case "testing":
		currentEnv = Testing
	default:
		currentEnv = Development
	}
	fmt.Printf("[ENV] Current environment: %s\n", currentEnv)
}

// IsDev 判断是否为开发环境
func IsDev() bool {
	return currentEnv == Development
}
