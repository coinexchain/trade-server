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

func NewPruneWorker(dir string) *PruneWorker {
	pw := &PruneWorker{}
	pw.work = msgqueue.NewFileDeleter(pw.doneHeightCh, dir)
	return pw
}

func (p *PruneWorker) Run() {
	p.work.Run()
	go p.tickHeight()
}

func (p *PruneWorker) tickHeight() {
	tick := time.NewTicker(60 * time.Second)
	defer tick.Stop()

	for {
		<-tick.C
		if dump, err := getHubDumpData(p.hub); err == nil && dump != nil {
			if dump.CurrBlockHeight != 0 {
				p.doneHeightCh <- dump.CurrBlockHeight
			}
		}
	}
}

func (p *PruneWorker) Close() {
	close(p.doneHeightCh)
}
