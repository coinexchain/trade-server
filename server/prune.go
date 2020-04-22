package server

import (
	"time"

	"github.com/coinexchain/cet-sdk/msgqueue"
	"github.com/coinexchain/trade-server/core"
)

type Worker interface {
	Run()
}

type WorkerCloser interface {
	Worker
	Close()
}

type PruneWorker struct {
	hub          *core.Hub
	doneHeightCh chan int64
	work         Worker
}

func NewPruneWorker(dir string, hub *core.Hub) *PruneWorker {
	pw := &PruneWorker{doneHeightCh: make(chan int64), hub: hub}
	pw.work = msgqueue.NewFileDeleter(pw.doneHeightCh, dir)
	return pw
}

func (p *PruneWorker) Run() {
	p.work.Run()
	go p.tickHeight()
}

func (p *PruneWorker) tickHeight() {
	tick := time.NewTicker(2 * time.Hour)
	defer tick.Stop()

	for {
		<-tick.C
		if dump, err := GetHubDumpData(p.hub); err == nil && dump != nil {
			if dump.CurrBlockHeight != 0 {
				p.doneHeightCh <- dump.CurrBlockHeight
			}
		}
	}
}

func (p *PruneWorker) Close() {
	close(p.doneHeightCh)
}
