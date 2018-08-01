package daemon

import "sync"

// CCController concurrent controller
type CCController struct {
	Wg     sync.WaitGroup
	CCChan chan struct{}
	Quit   chan struct{}
}

func NewCCController(ccNum int) *CCController {
	return &CCController{
		CCChan: make(chan struct{}, ccNum),
	}
}

func (cc *CCController) Add() {
	cc.CCChan <- struct{}{}
	cc.Wg.Add(1)
}

func (cc *CCController) Done() {
	<-cc.CCChan
	cc.Wg.Done()
}

func (cc *CCController) Wait() {
	cc.Wg.Wait()
}
