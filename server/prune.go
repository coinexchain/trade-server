package server

import (
	"time"

	"github.com/coinexchain/cet-sdk/msgqueue"
	"github.com/coinexchain/trade-server/core"
)

type Worker interface {
	Work()
}

type PruneWork struct {
	hub          *core.Hub
	doneHeightCh chan int64
	work         Worker
}

func NewPruneWork(dir string) *PruneWork {
	pw := &PruneWork{}
	pw.work = msgqueue.NewPruneFile(pw.doneHeightCh, dir)
	return pw
}

func (p *PruneWork) Work() {
	p.work.Work()
	go p.tickHeight()
}

func (p *PruneWork) tickHeight() {
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
