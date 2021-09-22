package task

import (
	"github.com/fanghongbo/log-agent/common/g"
	"github.com/fanghongbo/log-agent/handler"
	"github.com/robfig/cron"
	"sync"
)

var controller *Controller

func init() {
	controller = NewController()
}

type Callback struct {
	Cron        *cron.Cron `json:"cron"`
	Src         string     `json:"src"`
	Dest        string     `json:"dest"`
	Pattern     string     `json:"pattern"`
	Exclude     string     `json:"exclude"`
	ForbidWrite bool       `json:"forbidWrite"`
	AfterDelete bool       `json:"afterDelete"`
	BeforeTime  int64      `json:"beforeTime"`
	Sign        string     `json:"sign"`
	Spec        string     `json:"spec"`
}

func NewCallback(src string, dest string, pattern string, exclude string, forbidWrite bool, afterDelete bool, beforeTime int64, sign string, spec string) *Callback {
	return &Callback{Src: src, Dest: dest, Pattern: pattern, Exclude: exclude, ForbidWrite: forbidWrite, AfterDelete: afterDelete, BeforeTime: beforeTime, Sign: sign, Spec: spec}
}

type Controller struct {
	cbMap []*Callback
	quit  chan interface{}
	lock  sync.Mutex
}

func NewController() *Controller {
	return &Controller{cbMap: []*Callback{}, lock: sync.Mutex{}, quit: make(chan interface{}, 0)}
}

func (u *Controller) New(cb *Callback) {
	var (
		job *handler.LogService
		err error
	)

	u.lock.Lock()
	defer u.lock.Unlock()

	cb.Cron = cron.New()

	job = handler.NewLogService(cb.Src, cb.Dest, cb.Pattern, cb.Exclude, cb.ForbidWrite, cb.AfterDelete, cb.BeforeTime, cb.Sign, cb.Spec)

	if err = cb.Cron.AddJob(cb.Spec, job); err != nil {
		g.AppLog.Fatalf(err.Error())
	}

	u.cbMap = append(u.cbMap, cb)
}

func (u *Controller) Listen() {
	for _, item := range u.cbMap {
		if item.Cron == nil {
			continue
		}

		item.Cron.Start()
	}

	<-controller.quit // wait for quit
}

func (u *Controller) Stop() {
	for _, item := range u.cbMap {
		if item.Cron == nil {
			continue
		}

		item.Cron.Stop()
	}

	u.quit <- 1
}

func Start() {
	go func() {
		for _, item := range g.Cfg.Config().Strategies {
			controller.New(NewCallback(item.Src, item.Dest, item.Pattern, item.Exclude, item.ForbidWrite, item.AfterDelete, item.BeforeTime, item.Sign, item.Spec))
		}

		g.AppLog.Info("service start success")

		controller.Listen()
	}()
}

func Stop() {
	controller.Stop()
}
