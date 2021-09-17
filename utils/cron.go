package utils

import (
	"github.com/robfig/cron"
)

func CheckCronSpec(spec string) bool {
	var (
		c   *cron.Cron
		err error
	)

	c = cron.New()
	defer c.Stop()

	if err = c.AddFunc(spec, nil); err != nil {
		return false
	}

	c.Start()

	return true
}
