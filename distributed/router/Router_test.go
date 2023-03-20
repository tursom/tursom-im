package router

import (
	"context"
	"fmt"
	"testing"

	"github.com/tursom/GoCollections/exceptions"
)

type (
	testProcessor struct {
		id         string
		processors map[string]*testProcessor
		cluster    *routerImpl[any]
	}

	msgReq struct {
		target string
		msg    any
		jmp    uint32
	}

	msgResp struct {
		success bool
	}

	discoveryReq struct {
		target string
		jmp    uint32
	}

	discoveryResp struct {
		distance uint32
	}
)

func Test_cluster_Send(t1 *testing.T) {
	processors := make(map[string]*testProcessor)

	processors["1"] = &testProcessor{
		id:         "1",
		processors: processors,
	}
	processors["1"].cluster = NewRouter[any](DefaultConfig[any](processors["1"])).(*routerImpl[any])
	processors["1"].cluster.AddHosts([]string{"1", "2", "3"})
	processors["1"].cluster.SetDirectly([]string{"2"})

	processors["2"] = &testProcessor{
		id:         "2",
		processors: processors,
	}
	processors["2"].cluster = NewRouter[any](DefaultConfig[any](processors["2"])).(*routerImpl[any])
	processors["2"].cluster.AddHosts([]string{"1", "2", "3"})
	processors["2"].cluster.SetDirectly([]string{"1", "3"})

	processors["3"] = &testProcessor{
		id:         "3",
		processors: processors,
	}
	processors["3"].cluster = NewRouter[any](DefaultConfig[any](processors["3"])).(*routerImpl[any])
	processors["3"].cluster.AddHosts([]string{"1", "2", "3"})
	processors["3"].cluster.SetDirectly([]string{"2"})

	send, e := processors["1"].cluster.Send(context.Background(), "3", "hello", 0)
	fmt.Println(send, e)
	if e != nil {
		e.PrintStackTrace()
	}

	if !send || e != nil {
		t1.Error(e)
	}
}

func (p *testProcessor) send(ctx context.Context, msg any) (any, exceptions.Exception) {
	switch msg.(type) {
	case *msgReq:
		b := msg.(*msgReq)
		if b.target == p.id {
			fmt.Printf("[%d] %s received msg: %v\n", b.jmp, p.id, b.msg)
			return &msgResp{true}, nil
		} else {
			fmt.Printf("[%d] %s received msg request to %s: %v\n", b.jmp, p.id, b.target, b.msg)
			send, e := p.cluster.Send(ctx, b.target, b.msg, b.jmp)
			return &msgResp{send}, e
		}
	case *discoveryReq:
		req := msg.(*discoveryReq)
		if req.target == p.id {
			return &discoveryResp{1}, nil
		}

		distance, e := p.cluster.Find(ctx, req.target, req.jmp)

		if distance == UNREACHABLE {
			return &discoveryResp{distance}, e
		} else {
			return &discoveryResp{distance + 1}, e
		}
	default:
		return nil, exceptions.NewRuntimeException("unsupported msg", nil)
	}
}

func (p *testProcessor) LocalId() string {
	return p.id
}

func (p *testProcessor) Send(ctx context.Context, nextJmp string, target string, msg any, jmp uint32) (bool, exceptions.Exception) {
	fmt.Printf("%s send0(nextJmp: %s, terget: %s)\n", p.id, nextJmp, target)
	if target == p.id {
		fmt.Printf("%s received msg: %v\n", p.id, msg)
		return true, nil
	}

	nextJmpNode := p.processors[nextJmp]
	send, e := nextJmpNode.send(ctx, &msgReq{
		target: target,
		msg:    msg,
		jmp:    jmp,
	})

	fmt.Printf("%s send0(nextJmp: %s, terget: %s) success %v\n", p.id, nextJmp, target, send.(*msgResp).success)
	return send.(*msgResp).success, e
}

func (p *testProcessor) Find(ctx context.Context, nextJmp string, target string, jmp uint32) (uint32, exceptions.Exception) {
	fmt.Printf("%s find(nextJmp: %s, terget: %s)\n", p.id, nextJmp, target)
	if target == p.id {
		return 0, nil
	}

	processor := p.processors[nextJmp]
	find, e := processor.send(ctx, &discoveryReq{
		target: target,
		jmp:    jmp,
	})

	fmt.Printf("%s find(nextJmp: %s, terget: %s) distance %d\n", p.id, nextJmp, target, find.(*discoveryResp).distance)

	return find.(*discoveryResp).distance, e
}
