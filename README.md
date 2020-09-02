你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
# PROMISE
[![Go Report Card](https://goreportcard.com/badge/github.com/chebyrash/promise)](https://goreportcard.com/report/github.com/chebyrash/promise)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/chebyrash/promise)
[![Build Status](https://travis-ci.org/chebyrash/promise.svg?branch=master)](https://travis-ci.org/chebyrash/promise)

## About
Promises library for Golang. Inspired by [JS Promises.](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise)

## Installation

    $ go get -u github.com/chebyrash/promise

## Quick Start
```go
var p = promise.New(func(resolve func(interface{}), reject func(error)) {
    // Do something asynchronously.
    const sum = 2 + 2

    // If your work was successful call resolve() passing the result.
    if sum == 4 {
        resolve(sum)
        return
    }

    // If you encountered an error call reject() passing the error.
    if sum != 4 {
        reject(errors.New("2 + 2 doesnt't equal 4"))
        return
    }

    // If you forgot to check for errors and your function panics the promise will
    // automatically reject.
    // panic() == reject()
})

// A promise is a returned object to which you attach callbacks.
p.Then(func(data interface{}) interface{} {
    fmt.Println("The result is:", data)
    return data.(int) + 1
})

// Callbacks can be added even after the success or failure of the asynchronous operation.
// Multiple callbacks may be added by calling .Then or .Catch several times,
// to be executed independently in insertion order.
p.
    Then(func(data interface{}) interface{} {
        fmt.Println("The new result is:", data)
        return nil
    }).
    Catch(func(error error) error {
        fmt.Println("Error during execution:", error.Error())
        return nil
    })

// Since callbacks are executed asynchronously you can wait for them.
p.Await()
```

## Examples

### [HTTP Request](https://github.com/Chebyrash/promise/blob/master/examples/http_request/main.go)
```go
var requestPromise = promise.New(func(resolve func(interface{}), reject func(error)) {
    var url = "https://httpbin.org/ip"

    resp, err := http.Get(url)
    defer resp.Body.Close()

    if err != nil {
        reject(err)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        reject(err)
    }

    resolve(body)
})

requestPromise.
    // Parse JSON body
    Then(func(data interface{}) interface{} {
        var body = make(map[string]string)

        json.Unmarshal(data.([]byte), &body)

        return body
    }).
    // Work with parsed body
    Then(func(data interface{}) interface{} {
        var body = data.(map[string]string)

        fmt.Println("Origin:", body["origin"])

        return nil
    }).
    Catch(func(error error) error {
        fmt.Println(error.Error())
        return nil
    })

requestPromise.Await()
```

### [Finding Factorial](https://github.com/Chebyrash/promise/blob/master/examples/factorial/main.go)

```go
func findFactorial(n int) int {
	if n == 1 {
		return 1
	}
	return n * findFactorial(n-1)
}

func findFactorialPromise(n int) *promise.Promise {
	return promise.New(func(resolve func(interface{}), reject func(error)) {
		resolve(findFactorial(n))
	})
}

func main() {
	var factorial1 = findFactorialPromise(5)
	var factorial2 = findFactorialPromise(10)
	var factorial3 = findFactorialPromise(15)

	factorial1.Then(func(data interface{}) interface{} {
		fmt.Println("Result of 5! is", data)
		return nil
	})

	factorial2.Then(func(data interface{}) interface{} {
		fmt.Println("Result of 10! is", data)
		return nil
	})

	factorial3.Then(func(data interface{}) interface{} {
		fmt.Println("Result of 15! is", data)
		return nil
	})

	promise.AwaitAll(factorial1, factorial2, factorial3)
}
```

### [Chaining](https://github.com/Chebyrash/promise/blob/master/examples/http_request/main.go)
```go
var p = promise.New(func(resolve func(interface{}), reject func(error)) {
    resolve(0)
})

p.
    Then(func(data interface{}) interface{} {
        fmt.Println("I will execute first")
        return nil
    }).
    Then(func(data interface{}) interface{} {
        fmt.Println("And I will execute second!")
        return nil
    }).
    Then(func(data interface{}) interface{} {
        fmt.Println("Oh I'm last :(")
        return nil
    })

p.Await()
```
