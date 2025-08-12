package api

type RequestMethod int

const (
	GET RequestMethod = iota
	POST
	PUT
	DELETE
	HEAD
)

func (r RequestMethod) String() string {
	switch r {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case HEAD:
		return "HEAD"
	default:
		panic("Invalid request method")
	}
}

type Request struct {
	BaseURL         string                 ``
	Path            string                 ``
	Method          RequestMethod          ``
	InterfaceParams map[string]interface{} ``
	Params          map[string]string      ``
	FloatParams     map[string]float32     ``
	Float64Params   map[string]float64     ``
	IntParams       map[string]int         ``
	Int64Params     map[string]int64       ``
	HeaderField     map[string]string      ``
	FileField       map[string]string      ``
}

func NewRequest(baseURL string, path string, method RequestMethod) *Request {
	return &Request{
		BaseURL:         baseURL,
		Path:            path,
		Method:          method,
		InterfaceParams: map[string]interface{}{},
		Params:          map[string]string{},
		FloatParams:     map[string]float32{},
		Float64Params:   map[string]float64{},
		IntParams:       map[string]int{},
		Int64Params:     map[string]int64{},
		HeaderField:     map[string]string{},
		FileField:       map[string]string{},
	}
}
