package util

type Future struct {
	done chan bool
	val  interface{}
	err  error
}

func CreateFuture(loader func() (interface{}, error)) *Future {
	f := &Future{done: make(chan bool)}
	go func() {
		f.set(loader())
	}()
	return f
}

func (f *Future) set(val interface{}, err error) {
	f.val, f.err = val, err
	close(f.done)
}

func (f *Future) Get() (interface{}, error) {
	<-f.done
	return f.val, f.err
}
