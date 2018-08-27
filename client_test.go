package main

import (
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := simpleServer(ctx); err != nil {
			t.Fatalf("%s", err)
		}
	}()

	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
	if err != nil {
		t.Fatalf("can't connect to the proxy:%s", err)
	}
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	wg := sync.WaitGroup{}
	for i := 0; i < 300; i++ {
		go func(i int) {
			wg.Add(1)
			for j := 0; j < 10; j++ {
				req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:2014/%02d%02d", j, i), nil)
				if err != nil {
					t.Fatalf("can't create request:%s", err)
				}
				resp, err := httpClient.Do(req)
				if err != nil {
					t.Fatalf("can't GET page:%s", err)
				}
				if _, err := ioutil.ReadAll(resp.Body); err != nil {
					t.Fatalf("error reading body:%s", err)
				}
				resp.Body.Close()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	cancel()
}

func simpleServer(ctx context.Context) error {
	Mux := http.NewServeMux()
	Mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second)
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(request.URL.Path)[1:])
	})
	svr := http.Server{Addr: ":2014", Handler: Mux}
	if err := svr.ListenAndServe(); err != nil {
		return err
	}
	svr.Close()
	go svr.ListenAndServe()
	<-ctx.Done()
	svr.Close()
	return nil
}

func TestSocks5(t *testing.T) {

	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		t.Skipf("1080端口被占用，跳过本次测试")
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				conn, err := listener.Accept()
				if err != nil {
					t.Logf("error(%s):%s", reflect.TypeOf(err), err)
					continue
				}
				go handleConnection(conn)
			}
		}
	}()
	if !t.Run("", TestClient) {
		t.Fatalf("socks5代理存在错误")
	}
	cancel()
}
