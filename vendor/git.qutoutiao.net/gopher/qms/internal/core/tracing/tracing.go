package tracing

//const for tracing
const (
	HTTPMethod     = "http.method"
	HTTPPath       = "http.path"
	HTTPStatusCode = "http.status_code"
	HTTPHost       = "http.host"

	// PaaSProjectName tag
	PaaSProjectName = "paasProjectName"
)

type Option struct {
	FileName          string
	SamplingRate      string
	BufferSize        int64
	MaxTagValueLength int
}
