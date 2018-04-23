package main

import "context"

type Pipeline struct {
	channel chan *Record
	context context.Context
	cancel  context.CancelFunc
	err     error
}

func NewPipeline(ctx context.Context) *Pipeline {
	ctx, cancel := context.WithCancel(ctx)
	return &Pipeline{
		channel: make(chan *Record),
		context: ctx,
		cancel:  cancel,
		err:     nil,
	}
}

func (p *Pipeline) Send(record *Record) bool {
	select {
	case <-p.context.Done():
		close(p.channel)
		return false
	case p.channel <- record:
		return true
	}
}

func (p *Pipeline) Recv() <-chan *Record {
	select {
	case <-p.context.Done():
		close(p.channel)
		return p.channel
	default:
		return p.channel
	}
}

func (p *Pipeline) Wait() {
	<-p.context.Done()
}

func (p *Pipeline) Close() {
	p.cancel()
	close(p.channel)
}

func (p *Pipeline) Cancel() {
	p.cancel()
}

func (p *Pipeline) CancelWithError(err error) {
	select {
	case <-p.context.Done():
		return
	default:
		p.err = err
		p.cancel()
	}
}
