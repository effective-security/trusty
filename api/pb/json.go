package pb

import (
	"encoding/json"
)

// MarshalJSON implements json.Marshaler
func MarshalJSON(v any) ([]byte, error) {
	/*
		if msg, ok := v.(protoreflect.ProtoMessage); ok {
			return protojson.MarshalOptions{
				UseEnumNumbers:  false,
				EmitUnpopulated: false,
				UseProtoNames:   true,
				AllowPartial:    true,
				Multiline:       true,
				Indent:          "\t",
			}.Marshal(msg)
		}
	*/
	return json.MarshalIndent(v, "", "\t")
}
