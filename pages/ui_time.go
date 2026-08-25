package pages

import (
	"time"

	"github.com/dromara/carbon/v2"
)

func relativeTime(timestamp int64) string {
	if timestamp <= 0 {
		return "—"
	}
	if diff := time.Now().Unix() - timestamp; diff >= 0 && diff < 60 {
		return "刚刚"
	}
	return carbon.CreateFromTimestamp(timestamp).SetLocale("zh-CN").DiffForHumans()
}
