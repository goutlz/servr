package servr

import (
	"context"
	"net/http"
	"sync"

	"github.com/goutlz/errz"
)

type Server interface {
	Stop() error
}

type serverWrap struct {
	lock    sync.RWMutex
	server  *http.Server
	stopped bool
}

func (s *serverWrap) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.stopped {
		return errz.New("Server already stopped")
	}

	s.stopped = true

	err := s.server.Shutdown(context.Background())
	if err != nil {
		return errz.Wrap(err, "Failed to stop server")
	}

	return nil
}

func (s *serverWrap) isStopped() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.stopped
}

func New(addr string, router http.Handler) Server {
	wrap := &serverWrap{
		server: &http.Server{
			Handler: router,
			Addr:    addr,
		},
	}

	go func() {
		err := wrap.server.ListenAndServe()
		if err == nil {
			return
		}

		if wrap.isStopped() {
			return
		}

		panic(errz.Wrap(err, "ListenAndServe failed"))
	}()

	return wrap
}
