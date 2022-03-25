package httpsvc

import "net/http"

type Interceptor func(w http.ResponseWriter, r *http.Request, handler http.Handler)

type ServeMuxWrapper struct {
	*http.ServeMux
	chainedInterceptor Interceptor
}

func newServeMuxWrapper() *ServeMuxWrapper {
	return &ServeMuxWrapper{
		ServeMux: http.NewServeMux(),
	}
}

func (mux *ServeMuxWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mux.chainedInterceptor == nil {
		mux.ServeMux.ServeHTTP(w, r)
	} else {
		mux.chainedInterceptor(w, r, mux.ServeMux)
	}
}

func (mux *ServeMuxWrapper) Use(interceptors ...Interceptor) {
	if mux.chainedInterceptor == nil {
		mux.chainedInterceptor = chainServerInterceptor(interceptors...)
	} else {
		allInterceptors := append([]Interceptor{mux.chainedInterceptor}, interceptors...)
		mux.chainedInterceptor = chainServerInterceptor(allInterceptors...)
	}
}

func chainServerInterceptor(interceptors ...Interceptor) Interceptor {
	return func(w http.ResponseWriter, r *http.Request, handler http.Handler) {
		var state struct {
			i    int
			next http.Handler
		}
		state.next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if state.i == len(interceptors)-1 {
				interceptors[state.i](w, r, handler)
				return
			}

			state.i++
			interceptors[state.i-1](w, r, state.next)
		})
		state.next.ServeHTTP(w, r)
	}
}
