package generate

import (
	"bytes"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

const (
	content = `// Code generated by eventsource-protobuf. DO NOT EDIT.
// source: {{ .Source }}

package {{ .Package }}

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/altairsix/eventsource"
	"github.com/gogo/protobuf/proto"
)

type serializer struct {
}

func (s *serializer) MarshalEvent(event eventsource.Event) (eventsource.Record, error) {
	data, err := MarshalEvent(event)
	if err != nil {
		return eventsource.Record{}, err
	}

	return eventsource.Record{
		Version: event.EventVersion(),
		Data:    data,
	}, nil
}

func (s *serializer) UnmarshalEvent(record eventsource.Record) (eventsource.Event, error) {
	return UnmarshalEvent(record.Data)
}

func NewSerializer() eventsource.Serializer {
	return &serializer{}
}
{{ range .Fields }}
func (m *{{ .TypeName | base | camel }}) AggregateID() string { return m.{{ .TypeName | base | .ID }
func (m *{{ .TypeName | base | camel }}) EventVersion() int   { return int(m.Version) }
func (m *{{ .TypeName | base | camel }}) EventAt() time.Time  { return time.Unix(m.At, 0) }
{{ end }}

func MarshalEvent(event eventsource.Event) ([]byte, error) {
	container := &{{ .Message.Name | base | camel }}{}

	switch v := event.(type) {
{{ range .Fields }}
	case *{{ .TypeName | base | camel }}:
		container.Type = {{ .Number }}
		container.{{ .Name | camel }} = v
{{ end }}
	default:
		return nil, fmt.Errorf("Unhandled type, %v", event)
	}

	data, err := proto.Marshal(container)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func UnmarshalEvent(data []byte) (eventsource.Event, error) {
	container := &{{ .Message.Name | base | camel }}{};
	err := proto.Unmarshal(data, container)
	if err != nil {
		return nil, err
	}

	var event interface{}
	switch container.Type {
{{ range .Fields }}
	case {{ .Number }}:
		event = container.{{ .Name | camel }}
{{ end }}
	default:
		return nil, fmt.Errorf("Unhandled type, %v", container.Type)
	}

	return event.(eventsource.Event), nil
}

type Encoder struct{
	w io.Writer
}

func (e *Encoder) WriteEvent(event eventsource.Event) (int, error) {
	data, err := MarshalEvent(event)
	if err != nil {
		return 0, err
	}

	// Write the length of the marshaled event as uint64
	//
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, uint64(len(data)))
	if _, err := e.w.Write(buffer); err != nil {
		return 0, err
	}

	n, err := e.w.Write(data)
	if err != nil {
		return 0, err
	}

	return n + 8, nil
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

type Decoder struct {
	r       *bufio.Reader
	scratch *bytes.Buffer
}

func (d *Decoder) readN(n uint64) ([]byte, error) {
	d.scratch.Reset()
	for i := uint64(0); i < n; i++ {
		b, err := d.r.ReadByte()
		if err != nil {
			return nil, err
		}
		if err := d.scratch.WriteByte(b); err != nil {
			return nil, err
		}
	}
	return d.scratch.Bytes(), nil
}

func (d *Decoder) ReadEvent() (eventsource.Event, error) {
	data, err := d.readN(8)
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint64(data)

	data, err = d.readN(length)
	if err != nil {
		return nil, err
	}

	event, err := UnmarshalEvent(data)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder {
		r:       bufio.NewReader(r),
		scratch: bytes.NewBuffer(nil),
	}
}
`
)

// File accepts the proto file definition and returns the response for this file
func File(in *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {
	pkg, err := packageName(in)
	if err != nil {
		return nil, err
	}

	message, err := findContainerMessage(in)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	t, err := newTemplate(content)
	if err != nil {
		return nil, err
	}

	t.Execute(buf, map[string]interface{}{
		"Source":  *in.Name,
		"Package": pkg,
		"Message": message,
		"Fields":  message.Field[1:],
		"ID":      idFields(in),
	})

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    name(in),
		Content: String(buf.String()),
	}, nil
}

// AllFiles accepts multiple proto file definitions and returns the list of files
func AllFiles(in []*descriptor.FileDescriptorProto) ([]*plugin_go.CodeGeneratorResponse_File, error) {
	results := make([]*plugin_go.CodeGeneratorResponse_File, 0, len(in))

	if in != nil {
		for _, file := range in {
			v, err := File(file)
			if err != nil {
				return nil, err
			}
			results = append(results, v)
		}
	}

	return results, nil
}
