# handlers
Handlers is a small library of utility http handlers that are useful for building web applications. The library includes a NotFoundHandler, an ErrorHandler, and a FileHandler. 

NotFoundHandler and ErrorHandler provide a simple way to implement custom 404 and 500 status pages: create your own application specific error page templates and call each handler's Serve methods with the appropriate arguments.

FileHandler provides similar functionality to the FileServer in Go's [net/http][gnh] package, but with two differences: it will not show directory listings for directories under its path, and it will respond to any request for a non-existent file with the given NotFoundHandler.

### Installation
Install with `go get`.

```sh
go get github.com/olihawkins/handlers
```

### Tests
Use `go test` to run the tests.

### Documentation
See the [GoDoc][gd] for the full documentation.

### NotFoundHandler and ErrorHandler
To use the NotFoundHandler and the ErrorHandler, provide your own custom error templates when creating the handlers. The NotFoundHandler template should contain the {{.Path}} tag, while the ErrorHandler template should contain the {{.ErrorMessage}} tag. These handlers can be initialised in two ways, either by providing a path to the template file, or by providing a pointer to a struct of type Template from Go's [html/template][ght] package.
```go
// Create a NotFoundHandler with the given template file
templatePath := filepath.FromSlash("templates/notfound.html")
nfh := handlers.LoadNotFoundHandler(templatePath)

// Create a NotFoundHandler with the given *template.Template
nfh := handlers.NewNotFoundHandler(myNotFoundTemplate)

// Create an ErrorHandler with the given template file
templatePath := filepath.FromSlash("templates/error.html")
eh := handlers.LoadErrorHandler(templatePath, "Default error message", true)

// Create an ErrorHandler with the given *template.Template
eh := handlers.NewErrorHandler(myErrorTemplate, "Default error message", true)

```
As the above examples show, the functions used to create a NotFoundHandler only need a template, while the functions used to create an ErrorHandler take two more arguments. The first is a string that specifies the default error message to show when the handler's ServeError method is called. The second is a boolean that tells the handler whether to serve the default error message (false) or the specific error message passed to the ServeError method (true). This lets you report detailed error messages to the browser while developing, which can be turned-off later in production. The ErrorHandler's AlwaysServeError method lets you override the default error message even when the handler is set not to display specific errors.

These two handlers are intended to be used indirectly, from inside other handlers, where page not found or server errors occur and you need 
to report them to the browser. A simplified example http.Handler is shown below illustrating their use.

```go
// An example handler which uses a NotFoundHandler and an ErrorHandler
type ExampleHandler struct {
	eh *handlers.ErrorHandler
	nfh *handlers.NotFoundHandler
}

// ServeHTTP for the example handler
func (h *ExampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// If the request is not for the homepage, serve a 404
	if r.URL.Path != "/" {
		h.nfh.ServeHTTP(w, r)
		return
	}

	// Try to get data from a function that may return an error
	data, err := GetSomeData()

	// If the function returns an error, serve the error page
	if (err != nil) {
		h.eh.ServeError(w, "Could not get data.")
		return
	}

	fmt.Fprintf(w, "%s", data)
	return
}
```

### FileHandler
FileHandler provdes an alternative implementation of the default FileServer in Go's net/http package. Unlike the default FileServer, it does not show directory listings and will return 404 pages using the given NotFoundHandler.

```go
// Create a NotFoundHandler to use in the FileHandler
nfh := handlers.LoadNotFoundHandler("templates/notfound.html")

// Create a FileHandler for the "./test" directory and map it to path "/test/"
fh := handlers.NewFileHandler("/test/", "./test", nfh)
http.Handle("/test/", fh)
```
FileHandler's ServeHTTP method is just a wrapper around http.ServeFile, with paths and response values modified to provide the appropriate behaviour.

   [gd]: <https://godoc.org/github.com/olihawkins/handlers>
   [gnh]: <https://golang.org/pkg/net/http/>
   [ght]: <https://golang.org/pkg/html/template/>
