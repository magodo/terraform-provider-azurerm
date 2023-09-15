package types

import (
	"context"
	"sync"

	"github.com/hack-pad/safejs"
)

type listenable interface {
	MessageEventConnect | MessageEventMessage
}

// listen adds the EventListener on the listener for the specified events.
// It returns a channel, which will send the MessageEvent(s) listened on, until the ctx is canceled.
func listen[T listenable](ctx context.Context, listener safejs.Value, parseFunc func(safejs.Value) T, events ...string) (_ <-chan T, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	eventsCh := make(chan T)

	var handlers []safejs.Func
	var wg sync.WaitGroup
	for range events {
		handler, err := nonBlocking(&wg, func(args []safejs.Value) {
			select {
			case <-ctx.Done():
				return
			default:
				eventsCh <- parseFunc(args[0])
			}
		})
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	}

	go func() {
		<-ctx.Done()
		for i := range events {
			event, handler := events[i], handlers[i]
			_, err := listener.Call("removeEventListener", event, handler)
			if err == nil {
				handler.Release()
			}
		}
		wg.Wait()
		close(eventsCh)
	}()

	for i := range events {
		event, handler := events[i], handlers[i]
		_, err = listener.Call("addEventListener", event, handler)
		if err != nil {
			return nil, err
		}
	}

	return eventsCh, nil
}

func nonBlocking(wg *sync.WaitGroup, fn func(args []safejs.Value)) (safejs.Func, error) {
	return safejs.FuncOf(func(_ safejs.Value, args []safejs.Value) any {
		wg.Add(1)
		go func() {
			fn(args)
			wg.Done()
		}()
		return nil
	})
}
