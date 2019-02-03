package promise

import (
	"errors"
	"sync"
)

const (
	pending = iota
	fulfilled
	rejected
)

// A Promise is a proxy for a value not necessarily known when
// the promise is created. It allows you to associate handlers
// with an asynchronous action's eventual success value or failure reason.
// This lets asynchronous methods return values like synchronous methods:
// instead of immediately returning the final value, the asynchronous method
// returns a promise to supply the value at some point in the future.
type Promise struct {
	// A Promise is in one of these states:
	// Pending - 0. Initial state, neither fulfilled nor rejected.
	// Fulfilled - 1. Operation completed successfully.
	// Rejected - 2. Operation failed.
	state int

	// A function that is passed with the arguments resolve and reject.
	// The executor function is executed immediately by the Promise implementation,
	// passing resolve and reject functions (the executor is called
	// before the Promise constructor even returns the created object).
	// The resolve and reject functions, when called, resolve or reject
	// the promise, respectively. The executor normally initiates some
	// asynchronous work, and then, once that completes, either calls the
	// resolve function to resolve the promise or else rejects it if
	// an error or panic occurred.
	executor func(resolve func(interface{}), reject func(error))

	// Appends fulfillment to the promise,
	// and returns a new promise.
	then []func(data interface{}) interface{}

	// Appends a rejection handler to the promise,
	// and returns a new promise.
	catch []func(error error) error

	// Stores the result passed to resolve()
	result interface{}

	// Stores the error passed to reject()
	error error

	// Mutex protects against data race conditions.
	mutex *sync.Mutex

	// WaitGroup allows to block until all callbacks are executed.
	wg *sync.WaitGroup
}

// New instantiates and returns a *Promise object.
func New(executor func(resolve func(interface{}), reject func(error))) *Promise {
	var promise = &Promise{
		state:    pending,
		executor: executor,
		then:     make([]func(interface{}) interface{}, 0),
		catch:    make([]func(error) error, 0),
		result:   nil,
		error:    nil,
		mutex:    &sync.Mutex{},
		wg:       &sync.WaitGroup{},
	}

	go func() {
		defer promise.handlePanic()
		promise.executor(promise.resolve, promise.reject)
	}()

	return promise
}

func (promise *Promise) resolve(resolution interface{}) {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state != pending {
		return
	}

	promise.state = fulfilled
	promise.result = resolution
	doneCounter := 0

	for _, value := range promise.then {
		promise.result = value(promise.result)
		// check if returned value is promise
		if thenPromise, ok := promise.result.(*Promise); ok {
			isRejected := false

			thenPromise.Then(func(result interface{}) interface{} {
				promise.result = result
				return nil
			}).Catch(func(err error) error {
				chainError := err
				isRejected = true

				for i := 0; i < len(promise.then)-doneCounter; i++ {
					promise.wg.Done()
				}
				for _, value := range promise.catch {
					chainError = value(chainError)
					promise.wg.Done()
				}
				return chainError
			}).Await()

			if isRejected {
				return
			}
		}
		promise.wg.Done()
		doneCounter++
	}

	for range promise.catch {
		promise.wg.Done()
	}
}

func (promise *Promise) reject(error error) {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state != pending {
		return
	}

	for range promise.then {
		promise.wg.Done()
	}

	promise.error = error

	for _, value := range promise.catch {
		promise.error = value(promise.error)
		promise.wg.Done()
	}

	promise.state = rejected
}

func (promise *Promise) handlePanic() {
	var r = recover()
	if r != nil {
		promise.reject(errors.New(r.(string)))
	}
}

// Then appends fulfillment handler to the promise, and returns a new promise.
func (promise *Promise) Then(fulfillment func(data interface{}) interface{}) *Promise {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state == pending {
		promise.wg.Add(1)
		promise.then = append(promise.then, fulfillment)
	} else if promise.state == fulfilled {
		promise.result = fulfillment(promise.result)
	}

	return promise
}

// Catch appends a rejection handler callback to the promise, and returns a new promise.
func (promise *Promise) Catch(rejection func(error error) error) *Promise {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state == pending {
		promise.wg.Add(1)
		promise.catch = append(promise.catch, rejection)
	} else if promise.state == rejected {
		promise.error = rejection(promise.error)
	}

	return promise
}

// Await is a blocking function that waits for all callbacks to be executed.
func (promise *Promise) Await() {
	promise.wg.Wait()
}

// AwaitAll is a blocking function that waits for a number of promises to resolve
func AwaitAll(promises ...*Promise) {
	for _, promise := range promises {
		promise.Await()
	}
}
