package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/qerr"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var MaxConcurrent = errors.New("超过最多并发数量")

type Session struct {
	concurrent int
	counter    int
	quic.Session
	sync.Locker
	Addr string
	*tls.Config
}

func (s *Session) NewStream() (conn *Stream, err error) {
	if s.concurrent > 30 {
		return nil, MaxConcurrent
	}
	stream, err := s.OpenStream()
	if err == nil {
		if s.counter > 0 {
			s.Lock()
			s.counter = 0
			s.Unlock()
		}
		s.Lock()
		s.concurrent++
		s.Unlock()
		return &Stream{s, stream, &sync.WaitGroup{}}, nil
	}
	if s.counter > 3 {
		return nil, err
	}
	s.Lock()
	s.counter++
	s.Unlock()
	if QuicError, ok := err.(*qerr.QuicError); !ok || QuicError.ErrorCode != qerr.PublicReset {
		return nil, err
	}
	s.Session, err = quic.DialAddr(s.Addr, s.Config, &quic.Config{KeepAlive: true})
	if err == nil {
		return s.NewStream()
	}
	return nil, err
}

type Stream struct {
	*Session
	quic.Stream
	*sync.WaitGroup
}

func (s *Stream) Close() error {
	s.Lock()
	defer s.Unlock()
	s.concurrent--
	s.WaitGroup.Wait()
	return s.Stream.Close()
}

func (s *Stream) Write(p []byte) (n int, err error) {
	s.WaitGroup.Add(1)
	n, err = s.Stream.Write(p)
	s.Done()
	return
}

type Manager struct {
	ss []*Session
	sync.Locker
	Addr string
	*tls.Config
}

func NewManager(addr string, config *tls.Config) (*Manager, error) {
	m := &Manager{ss: make([]*Session, 0), Locker: &sync.Mutex{}, Addr: addr, Config: config}
	err := m.NewSession()
	if err != nil {
		return nil, err
	}
	m.Deamon()
	return m, nil
}
func (m *Manager) Deamon() {
	go func() {
		for {
			time.Sleep(time.Minute)
			i := 0
			for m.Recycling() {
				i++
			}
			if i > 0 {
				logf("清理 %d 个空闲的会话,当前剩余 %d", i,len(m.ss))
			}
		}
	}()
}
func (m *Manager) Recycling() bool {
	for i := range m.ss {
		if m.ss[i].concurrent == 0 {
			m.Lock()
			m.Unlock()
			m.ss[i].Session.Close()
			m.ss = append(m.ss[0:i], m.ss[i+1:]...)
			return true
		}
	}
	return false
}
func (m *Manager) NewSession() error {
	s, err := quic.DialAddr(m.Addr, m.Config, nil)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	m.ss = append(m.ss, &Session{concurrent: 0, counter: 0, Session: s, Locker: &sync.Mutex{}, Addr: m.Addr, Config: m.Config})
	return nil
}
func (m *Manager) NewStream() (stream *Stream, err error) {
	if len(m.ss) == 0 {
		err := m.NewSession()
		if err != nil {
			return nil, err
		}
	}
	session := m.ss[len(m.ss)-1]
	stream, err = session.NewStream()
	if err == nil {
		return
	}
	if err == MaxConcurrent {
		err := m.NewSession()
		if err != nil {
			return nil, err
		}
		return m.NewStream()
	}
	return nil, err
}

func logf(format string, a ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("\x1B[01;33m%s %s[%d]:\x1B[0m",
		time.Now().Format("15:04:05"),
		filepath.Base(file),
		line,
	)
	fmt.Printf(format, a...)
	fmt.Println()
}
