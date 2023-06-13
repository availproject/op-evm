package avail

import (
	"net/http"
	"sync"
)

type FraudServer struct {
	mutex   *sync.Mutex
	fraudFn *sync.Once
}

func NewFraudServer() *FraudServer {
	s := &FraudServer{
		mutex:   new(sync.Mutex),
		fraudFn: new(sync.Once),
	}

	// Dispose first fn invocation.
	s.fraudFn.Do(func() {})
	return s
}

func (fs *FraudServer) PerformFraud(f func()) {
	fs.mutex.Lock()
	fs.fraudFn.Do(f)
	fs.mutex.Unlock()
}

func (fs *FraudServer) PrimeFraud() {
	fs.mutex.Lock()
	fs.fraudFn = new(sync.Once)
	fs.mutex.Unlock()
}

func (fs *FraudServer) ListenAndServe(addr string) error {
	http.HandleFunc("/fraud/prime", func(w http.ResponseWriter, _ *http.Request) {
		fs.PrimeFraud()
		w.WriteHeader(http.StatusAccepted)
	})
	return http.ListenAndServe(addr, nil)
}
