package sse_test

import (
	"sync"
	"testing"
	"time"

	"mdserve/internal/sse"
)

// --- Task 6.2: SSEBroker ---

func TestBroker_RegisterReturnsChannel(t *testing.T) {
	b := sse.New()
	ch := b.Register()
	if ch == nil {
		t.Fatal("Register() returned nil channel")
	}
}

func TestBroker_BroadcastSendsToRegisteredClient(t *testing.T) {
	b := sse.New()
	ch := b.Register()

	b.Broadcast()

	select {
	case <-ch:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Error("Broadcast() did not send to registered client")
	}
}

func TestBroker_BroadcastSendsToMultipleClients(t *testing.T) {
	b := sse.New()
	ch1 := b.Register()
	ch2 := b.Register()
	ch3 := b.Register()

	b.Broadcast()

	for i, ch := range []<-chan struct{}{ch1, ch2, ch3} {
		select {
		case <-ch:
			// success
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Broadcast() did not send to client %d", i+1)
		}
	}
}

func TestBroker_UnregisterRemovesClient(t *testing.T) {
	b := sse.New()
	ch := b.Register()
	b.Unregister(ch)

	b.Broadcast()

	// 登録解除後はチャンネルに何も届かないこと
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Unregistered client received a broadcast event")
		}
		// closed channel (ok==false) is expected
	case <-time.After(50 * time.Millisecond):
		// no event: also expected (channel closed or never sent)
	}
}

func TestBroker_BroadcastIsNonBlocking(t *testing.T) {
	b := sse.New()
	// バッファなしチャンネルのクライアントを登録してもBroadcastがブロックしないこと
	_ = b.Register() // チャンネルを読まない

	done := make(chan struct{})
	go func() {
		b.Broadcast()
		close(done)
	}()

	select {
	case <-done:
		// success: Broadcast returned immediately
	case <-time.After(200 * time.Millisecond):
		t.Error("Broadcast() blocked when client channel is full")
	}
}

// --- Task 1.1: Broker.Shutdown() ---

func TestBroker_ShutdownClosesAllChannels(t *testing.T) {
	b := sse.New()
	ch1 := b.Register()
	ch2 := b.Register()

	b.Shutdown()

	for i, ch := range []<-chan struct{}{ch1, ch2} {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("client %d channel not closed after Shutdown()", i+1)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client %d channel not closed after Shutdown() (timeout)", i+1)
		}
	}
}

func TestBroker_ShutdownWithNoClients(t *testing.T) {
	b := sse.New()
	// クライアントなしでShutdown()を呼んでもパニックしないこと
	b.Shutdown()
}

func TestBroker_ConcurrentRegisterUnregisterBroadcast(t *testing.T) {
	b := sse.New()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := b.Register()
			b.Broadcast()
			b.Unregister(ch)
		}()
	}

	// デッドロックやパニックなく完了すること
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("concurrent Register/Unregister/Broadcast timed out (possible deadlock)")
	}
}
