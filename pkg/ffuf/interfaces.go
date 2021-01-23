package ffuf

//FilterProvider is a generic interface for both Matchers and Filters
type FilterProvider interface {
	Filter(response *Response) (bool, error)
	Repr() string
}

//RunnerProvider is an interface for request executors
type RunnerProvider interface {
	Prepare(input map[string][]byte) (Request, error)
	Execute(req *Request) (Response, error)
}

//InputProvider interface handles the input data for RunnerProvider
type InputProvider interface {
	AddProvider(InputProviderConfig) error
	Next() bool
	Position() int
	Reset()
	Value() map[string][]byte
	Total() int
	Inject(keyword string, values [][]byte)
}

//InternalInputProvider interface handles providing input data to InputProvider
type InternalInputProvider interface {
	Keyword() string
	Next() bool
	Position() int
	ResetPosition()
	IncrementPosition()
	Value() []byte
	Total() int
	Inject(values [][]byte)
}

//OutputProvider is responsible of providing output from the RunnerProvider
type OutputProvider interface {
	Banner() error
	Finalize() error
	Progress(status Progress)
	Info(infostring string)
	Error(errstring string)
	Warning(warnstring string)
	Result(resp Response)
}
