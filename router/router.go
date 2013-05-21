package router

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// HTTP 1.1 Methods
const (
	DELETE = "DELETE"
	GET    = "GET"
	PATCH  = "PATCH"
	POST   = "POST"
	PUT    = "PUT"
)

//mime-types
const (
	applicationJSON = "application/json"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler http.HandlerFunc
}

type Router struct {
	routes  []*route
	filters []http.HandlerFunc
}

func New() *Router {
	return &Router{}
}

// Request-URI method implementations
func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.AddRoute(GET, pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.AddRoute(PUT, pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.AddRoute(DELETE, pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.AddRoute(PATCH, pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.AddRoute(POST, pattern, handler)
}

func (r *Router) AddRoute(method string, pattern string, handler http.HandlerFunc) {
	parts := strings.Split(pattern, "/")
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			// a user can override the defult expression
			// eg: ‘/cats/:id([0-9]+)’
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
			}
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	pattern = strings.Join(parts, "/")
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		// TODO: avoid panic
		panic(regexErr)
		return
	}

	route := &route{}
	route.method = method
	route.regex = regex
	route.handler = handler
	route.params = params
	r.routes = append(r.routes, route)
}

// required by http.Handler interface
// matches request with a route, and if found, serves the request with the route's handler
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestPath := req.URL.Path
	w := &responseWriter{writer: rw}

	for _, route := range r.routes {
		if req.Method != route.method || !route.regex.MatchString(requestPath) {
			continue
		}

		//get param submatches
		matches := route.regex.FindStringSubmatch(requestPath)

		//check that route matches URL pattern
		if len(matches[0]) != len(requestPath) {
			continue
		}

		if len(route.params) > 0 {
			//push URL params to query param map
			values := req.URL.Query()
			for i, match := range matches[1:] {
				values.Add(route.params[i], match)
			}

			//concatenate query params and RawQuery
			req.URL.RawQuery = url.Values(values).Encode() + "&" + req.URL.RawQuery
			//req.URL.RawQuery = url.Values(values).Encode()
		}

		//call each middleware filter
		for _, filter := range r.filters {
			filter(w, req)
			if w.started {
				return
			}
		}
		route.handler(w, req)
		break
	}

	//return http.NotFound if no route matches the request
	if w.started == false {
		http.NotFound(w, req)
	}
}

type responseWriter struct {
	writer  http.ResponseWriter
	started bool
	status  int
}

func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.started = true
	return w.writer.Write(p)
}
