package job

import (
	"context"
	"kratos-realworld/internal/biz/profile"
	"time"
)

func StartRepairJob(repo *profile.ProfileRepo) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			repo.RepairFollowCache(context.Background())
		}
	}()
}
