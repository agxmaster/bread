package render

import (
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type Jsonpb struct {
	Data interface{}
}

var jsonpbContentType = []string{"application/json; charset=utf-8"}

// Render (Jsonpb) marshals the given interface object and writes data with custom ContentType.
func (r Jsonpb) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	marshaler := jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	if err := marshaler.Marshal(w, r.Data.(proto.Message)); err != nil {
		return err
	}
	return nil
}

// WriteContentType (Jsonpb) writes ProtoBuf ContentType.
func (r Jsonpb) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonpbContentType)
}
