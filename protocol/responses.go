package protocol

//TODO: Handle errors

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"github.com/olsdavis/goelan/util"
	"github.com/olsdavis/goelan/world"
)

type ChatComponent struct {
	Text string `json:"text"`
}

type Response struct {
	data *bytes.Buffer
}

// NewResponse creates a new response.
func NewResponse() *Response {
	return &Response{data: new(bytes.Buffer)}
}

// WriteBoolean writes the given boolean to the current response.
func (r *Response) WriteBoolean(b bool) *Response {
	if b {
		return r.WriteByte(1)
	} else {
		return r.WriteByte(0)
	}
}

// WriteByte writes the given byte to the current response.
func (r *Response) WriteByte(b int8) *Response {
	r.data.Write([]byte{byte(b)})
	return r
}

// WriteUnsignedByte writes the given unsigned byte to the current response.
func (r *Response) WriteUnsignedByte(b uint8) *Response {
	binary.Write(r.data, ByteOrder, b)
	return r
}

// WriteUVarint writes the given UVarint to the current response.
func (r *Response) WriteUVarint(i uint32) *Response {
	_, err := r.data.Write(Uvarint(i))
	if err != nil {
		panic(err)
	}
	return r
}

// WriteVarint writes the given Varint to the current response.
func (r *Response) WriteVarint(i int32) *Response {
	_, err := r.data.Write(Varint(i))
	if err != nil {
		panic(err)
	}
	return r
}

// WriteInt writes the given integer to the current response.
func (r *Response) WriteInt(i int) *Response {
	binary.Write(r.data, ByteOrder, int32(i))
	return r
}

// WriteLong writes the given long to the current response.
func (r *Response) WriteLong(l int64) *Response {
	binary.Write(r.data, ByteOrder, l)
	return r
}

// WriteFloat writes the given float to the current response.
func (r *Response) WriteFloat(f float32) *Response {
	binary.Write(r.data, ByteOrder, f)
	return r
}

// WriteDouble writes the given double to the current response.
func (r *Response) WriteDouble(d float64) *Response {
	binary.Write(r.data, ByteOrder, d)
	return r
}

// WriteUUID writes the most and then the least significant
// bits of the given UUID.
func (r *Response) WriteUUID(uuid util.UUID) {
	r.WriteLong(uuid.MostSig)
	r.WriteLong(uuid.LeastSig)
}

// WriteByteArray writes the given byte array to the current response.
func (r *Response) WriteByteArray(b []byte) *Response {
	r.WriteUVarint(uint32(len(b)))
	r.data.Write(b)
	return r
}

// WriteString writes the given string to the current response.
func (r *Response) WriteString(str string) *Response {
	return r.WriteByteArray([]byte(str))
}

// WriteJSON writes the given interface as a JSON string to the current response.
// (Currently ignores failure.)
func (r *Response) WriteJSON(obj interface{}) *Response {
	j, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return r.WriteByteArray(j)
}

// WriteStructure explores the field of the given interface, and, if
// their serialization is supported, writes them to the response;
// otherwise, tries to explore the field as an interface too--if
// it fails, skips the field.
func (r *Response) WriteStructure(object interface{}) *Response {
	switch object.(type) {
	case int8:
		r.WriteByte(object.(int8))
	case uint8:
		r.WriteUnsignedByte(object.(byte))
	case bool:
		r.WriteBoolean(object.(bool))
	case int:
		r.WriteInt(object.(int))
	case int32:
		r.WriteVarint(object.(int32))
	case int64:
		r.WriteLong(object.(int64))
	case uint32:
		r.WriteUVarint(object.(uint32))
	case float32:
		r.WriteFloat(object.(float32))
	case float64:
		r.WriteDouble(object.(float64))
	case string:
		r.WriteString(object.(string))
	case []byte:
		r.WriteByteArray(object.([]byte))
	case world.Location:
		//TODO: fix
		loc := object.(world.Location)
		r.WriteDouble(float64(loc.X))
		r.WriteDouble(float64(loc.Y))
		r.WriteDouble(float64(loc.Z))
		r.WriteFloat(loc.Yaw)
		r.WriteFloat(loc.Pitch)
	case util.UUID:
		r.WriteUUID(object.(util.UUID))
	default:
		t := reflect.ValueOf(object)
		if t.CanInterface() {
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Kind() == reflect.Ptr {
					if f.IsNil() {
						continue
					}
					f = f.Elem()
				}
				// ignore nil values
				if f.Kind() == reflect.Struct && reflect.Zero(f.Type()) == f {
					continue
				}
				r.WriteStructure(f.Interface())
			}
		}
	}
	return r
}

// ToRawPacket creates a raw packet from the written bytes and the provided id.
func (r *Response) ToRawPacket(id uint64) *RawPacket {
	return NewRawPacket(id, r.data.Bytes(), nil)
}

// Clear clears the data from the response's buffer.
func (r *Response) Clear() {
	r.data = new(bytes.Buffer)
}
