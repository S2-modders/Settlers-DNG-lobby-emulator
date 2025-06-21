package packages

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	ser "github.com/kelindar/binary"
	"golang.org/x/text/encoding/charmap"
)

func Stringify(source any) string {
	var builder strings.Builder

	v := reflect.ValueOf(source).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := v.Type().Field(i)

		if ! f.CanInterface() {
			continue
		}

		var str string

		switch v := f.Interface().(type) {
		case string:
			str = fmt.Sprintf(
				"%s: string(%d) | %s\n",
				ft.Name, len(v), v,
			)
		case []byte:
			str = fmt.Sprintf(
				"%s: bytes(%d) | %X\n",
				ft.Name, len(v), v,
			)
		default:
			str = fmt.Sprintf(
				"%s: %T | %v\n",
				ft.Name, v, v,
			)
		}

		builder.WriteString("\t")
		builder.WriteString(str)
	}

	return builder.String()
}

func Serialize(w io.Writer, source any) error {
	if k := reflect.TypeOf(source).Kind(); k != reflect.Pointer {
		return fmt.Errorf("serialize only supports pointers: %s", k.String())
	}

	writer := ser.NewEncoder(w)
	v := reflect.ValueOf(source).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		if ! f.CanInterface() {
			return fmt.Errorf("cannot interface field of target")
		}

		switch v := f.Interface().(type) {
		case uint8:
			writer.Write([]byte{v})
		case uint16:
			writer.WriteUint16(v)
		case uint32:
			writer.WriteUint32(v)
		case bool:
			if v {
				writer.Write([]byte{1})
			} else {
				writer.Write([]byte{0})
			}
		case string:
			if err := writeString(writer, v); err != nil {
				return err
			}
		case []byte:
			writer.WriteUint32(uint32(len(v)))
			writer.Write(v)
		default:
			return fmt.Errorf("field of interface has unknown type: %s", f.Kind().String())
		}

	}

	return nil
}

func Deserialize(r io.Reader, target any) error {
	if k := reflect.TypeOf(target).Kind(); k != reflect.Pointer {
		return fmt.Errorf("deserialize only supports pointers: %s", k.String())
	}

	reader := ser.NewDecoder(r)
	v := reflect.ValueOf(target).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)

		if ! f.CanInterface() {
			return fmt.Errorf("cannot interface field of target")
		}
		if ! f.IsValid() || ! f.CanSet() {
			return fmt.Errorf("field of interface is not settable")
		}

		switch f.Interface().(type) {
		case uint8:
			val := make([]byte, 1)
			if _, err := reader.Read(val); err != nil {
				return err
			}
			f.SetUint(uint64(val[0]))
		case uint16:
			val, err := reader.ReadUint16()
			if err != nil {
				return err
			}
			f.SetUint(uint64(val))
		case uint32:
			val, err := reader.ReadUint32()
			if err != nil {
				return err
			}
			f.SetUint(uint64(val))
		case bool:
			val := make([]byte, 1)
			if _, err := reader.Read(val); err != nil {
				return err
			}
			f.SetBool(val[0] > 0)
		case string:
			val, err := readString(reader)
			if err != nil {
				return err
			}
			f.SetString(val)
		case []byte:
			val, err := readSlice(reader)
			if err != nil {
				return err
			}
			f.SetBytes(val)
		default:
			return fmt.Errorf("field of interface has unknown type: %s", f.Kind().String())
		}

	}

	return nil
}


// NOTE: different versions of games might use other encoding
var stringDecoder = charmap.ISO8859_15.NewDecoder()
var stringEncoder = charmap.ISO8859_15.NewEncoder()


func readString(r *ser.Decoder) (string, error) {
	length, err := r.ReadUint32()
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}

	s, err := stringDecoder.String(ser.ToString(&buf))
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(s, "\x00"), nil
}

func writeString(w *ser.Encoder, s string) error {
	if ! strings.HasSuffix(s, "\x00") {
		s = s + "\x00"
	}

	s, err := stringEncoder.String(s)
	if err != nil {
		return err
	}

	w.WriteUint32(uint32(len(s)))
	w.Write(ser.ToBytes(s))
	
	return nil
}


func readSlice(r *ser.Decoder) ([]byte, error) {
	length, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	return buf, nil
}
