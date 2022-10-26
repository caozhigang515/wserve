package wserve

import "net/http"

type options struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
	ReadDeadline    int
	WriteDeadline   int
	MaxMessageSize  int64
	Certification   func(*http.Request) (IUser, error)
	DeBug           bool
	Permissions     func(*http.Request, string) bool
}

func DefaultOptions() *options {
	return &options{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadDeadline:   60,
		WriteDeadline:  10,
		MaxMessageSize: 512,
		Certification: func(request *http.Request) (IUser, error) {
			return &DUser{Addr: request.RemoteAddr}, nil
		},
	}
}

type Option func(*options)

func SetReadBufferSize(v int) Option {
	return func(o *options) {
		o.ReadBufferSize = v
	}
}

func SetWriteBufferSize(v int) Option {
	return func(o *options) {
		o.WriteBufferSize = v
	}
}

func SetCheckOrigin(fn func(r *http.Request) bool) Option {
	return func(o *options) {
		o.CheckOrigin = fn
	}
}

func SetReadDeadline(v int) Option {
	return func(o *options) {
		o.ReadDeadline = v
	}
}

func SetWriteDeadline(v int) Option {
	return func(o *options) {
		o.WriteDeadline = v
	}
}

func SetMaxMessageSize(v int64) Option {
	return func(o *options) {
		o.MaxMessageSize = v
	}
}

func SetCertification(fn func(*http.Request) (IUser, error)) Option {
	return func(o *options) {
		o.Certification = fn
	}
}

func Debug() Option {
	return func(o *options) {
		o.DeBug = true
	}
}

func SetPermissions(fn func(*http.Request, string) bool) Option {
	return func(o *options) {
		o.Permissions = fn
	}
}
