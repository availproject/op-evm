package future

import "sync"

// Future type provides generic Future object.
type Future[T any] struct {
	valueSet bool
	value    T
	err      error
	once     *sync.Once
	wg       *sync.WaitGroup
}

// New instantiates a Future
func New[T any]() Future[T] {
	f := Future[T]{}
	f.once = &sync.Once{}
	f.wg = &sync.WaitGroup{}
	f.wg.Add(1)
	return f
}

func (f *Future[T]) SetError(err error) {
	if f.valueSet {
		panic("value already set")
	}

	f.once.Do(func() {
		f.err = err
		f.valueSet = true
		f.wg.Done()
	})
}

// SetValue sets the return value for Future.
func (f *Future[T]) SetValue(v T) {
	if f.valueSet {
		panic("value already set")
	}

	f.once.Do(func() {
		f.value = v
		f.valueSet = true
		f.wg.Done()
	})
}

// Ready returns boolean, whether Future result is ready yet or not.
func (f *Future[T]) Ready() bool {
	return f.valueSet
}

// Result returns the result value of Future. It blocks until result becomes available.
func (f *Future[T]) Result() (T, error) {
	f.wg.Wait()
	return f.value, f.err
}
