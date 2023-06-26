package avail

import (
	"net/http"
	"sync"
)

// FraudServer is a server for managing and performing fraud detection operations.
// It uses a mutex for synchronization and a sync.Once to ensure fraud detection is performed exactly once per invocation.
type FraudServer struct {
	mutex   *sync.Mutex // mutex is used to lock and unlock the server during critical operations.
	fraudFn *sync.Once  // fraudFn is used to ensure a fraud detection operation is performed only once.
}

// NewFraudServer creates a new instance of FraudServer with the mutex and fraudFn initialized.
// It returns a pointer to the new instance of FraudServer.
// The first invocation of fraudFn is disposed off immediately to make the FraudServer ready for subsequent uses.
func NewFraudServer() *FraudServer {
	s := &FraudServer{
		mutex:   new(sync.Mutex),
		fraudFn: new(sync.Once),
	}

	// Dispose first fn invocation.
	s.fraudFn.Do(func() {})
	return s
}

// PerformFraud takes a function as an argument and performs the fraud detection operation by calling the function exactly once.
// The function is performed under a mutex lock to ensure thread-safety.
func (fs *FraudServer) PerformFraud(f func()) {
	fs.mutex.Lock()
	fs.fraudFn.Do(f)
	fs.mutex.Unlock()
}

// PrimeFraud resets the fraudFn to make it ready for the next invocation of fraud detection operation.
// It is performed under a mutex lock to ensure thread-safety.
func (fs *FraudServer) PrimeFraud() {
	fs.mutex.Lock()
	fs.fraudFn = new(sync.Once)
	fs.mutex.Unlock()
}

// ListenAndServe starts the FraudServer and listens for incoming HTTP requests on the specified address.
// It sets up a HTTP handler at "/fraud/prime" to prime the fraud detection operation for the next invocation.
func (fs *FraudServer) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/fraud/prime", func(w http.ResponseWriter, _ *http.Request) {
		fs.PrimeFraud()
		w.WriteHeader(http.StatusAccepted)
	})
	return http.ListenAndServe(addr, mux)
}
