JSONError for Golang [![GoDoc](http://godoc.org/github.com/pjebs/jsonerror?status.svg)](http://godoc.org/github.com/pjebs/jsonerror) 
=============

This package is for adding some structure to your error messages. This makes error-handling, debugging and diagnosis for all your Go projects a lot more elegant and simpler. Use it wherever `error` type is required.

It utilizes the fact that built-in [`type error`](http://golang.org/pkg/builtin/#error) is actually an `interface.`


Since Go is a new programming language, I have made the documentation and code as easy to understand as possible. Studying the code can be a great learning experience.

Install
-------

```shell
go get -u github.com/pjebs/jsonerror
```

Optional - if you want to output JSON formatted error messages (e.g. for REST API):

```shell
go get -u gopkg.in/unrolled/render.v1
```

Prehistoric Usage - Using [Go Standard Library](http://golang.org/pkg/errors/#example_New)
-----

```go
import (
	"errors"
	"fmt"
)

func main() {
	err := errors.New("emit macho dwarf: elf header corrupted")
	if err != nil {
		fmt.Print(err)
	}
}

//Or alternatively

panic(errors.New("failed"))

```


Using this package instead
-----

```go
import (
	e "github.com/pjebs/jsonerror" //aliased for ease of usage
	"math"                         //For realSquareRoot() example function below
)

//EXAMPLE 1 - Creating a JE Struct

err := e.New(1, "Square root of negative number is prohibited", "Please make number positive or zero") //Domain is optional and not included here

//Or  
err := e.New(1, "Square root of negative number is prohibited", "Please make number positive or zero", "com.github.pjebs.jsonerror")

//EXAMPLE 2 - Practical Example

//Custom function
func realSquareRoot(n float64) (float64, error) {
	if n < 0 {
		return 0, e.New(1, "Square root of negative number is prohibited", "Please make number positive or zero")
	} else {
		return math.Sqrt(n), nil
	}
}

//A function that uses realSquareRoot
func main() {

	s, err := realSquareRoot(12.0)
	if err == nil {
		//s is Valid answer
	} else {
		if err.(e.JE).Code == 1 {
			//Square root of negative number
		} else {
			//Unknown error
		}
	}
}


```

Methods
--------

```go
func New(code int, error string, message string, domain ...string) JE
```

`code int` - Error code. Arbitrary and set by *fiat*. Different types of errors should have an unique `error code` in your project.

`error string` - A standard description of the `error code.`

`message string` - A more detailed description that may be customized for the particular circumstances. May also provide extra information.

`domain ...string` - *Optional* It allows you to distinguish between same `error codes.` Only 1 `domain` string is allowed.


```go
func (this JE) Render() map[string]string {
```

Formats `JE` (JSONError) struct so it can be used by [gopkg.in/unrolled/render.v1](https://github.com/unrolled/render) package to generate JSON output.


Output JSON formatted error message (i.e. REST API Server response)
----------

```go
import (
	"github.com/codegangsta/negroni" //Using Negroni (https://github.com/codegangsta/negroni)
	e "github.com/pjebs/jsonerror"
	"gopkg.in/unrolled/render.v1"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		err := e.New(12, "Unauthorized Access", "Please log in first to access this site")

    	r := render.New(render.Options{})
		r.JSON(w, http.StatusUnauthorized, err.Render())
		return
  	
  	})

  	n := negroni.Classic()
  	n.UseHandler(mux)
  	n.Run(":3000")
}

```

For the above example, the web server will respond with a HTTP Status Code of 401 (Status Unauthorized), Content-Type as application/json and a JSON response:

```json
{"code":"12","error":"Unauthorized Access","message":"Please log in first to access this site"}
```

FAQ
--------

**What is the domain parameter?**

The domain parameter is optional. It allows you to distinguish between same error codes. That way different packages (or different parts of your own project) can use the same error codes (for different purposes) and still be differentiated by the domain identifier.

NB: The domain parameter is not outputted by `Render()` (for generating JSON formatted output)

**How do I use this package?**

When you want to return an error (e.g. from a function), just return a `JE` struct. See the example code above.

Or you can use it with `panic()`.

```go
panic(jsonerror.New(1, "error", "message"))
```

**What are the error codes?**

You make them up for your particular project! By *fiat*, you arbitrarily set each error code to mean a different *type* of error.


Final Notes
--------

If you found this package useful, please **Star** it on github. Feel free to fork or provide pull requests. Any bug reports will be warmly received.


[PJ Engineering and Business Solutions Pty. Ltd.](http://www.pjebs.com.au)
