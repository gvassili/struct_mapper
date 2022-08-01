package struct_mapper

import (
	"fmt"
	"reflect"
	"sync"
)

type HandleFunc func(v interface{}) interface{}

type Mapper struct {
	decoders sync.Map
	handlers map[string]HandleFunc
}

func New() *Mapper {
	return &Mapper{
		handlers: make(map[string]HandleFunc),
	}
}

func (m *Mapper) GetDecoder(src interface{}, dst interface{}) (*StructMapDecoder, error) {
	srcT := reflect.TypeOf(src).Elem()
	dstT := reflect.TypeOf(dst).Elem()
	decName := fmt.Sprint(srcT.Name(), "#", dstT.Name())
	tmpl, ok := m.decoders.Load(decName)
	if !ok {
		newTmpl, err := newStructMapDecoder(srcT, dstT, m.handlers)
		if err != nil {
			return nil, fmt.Errorf("creating writer for type %s: %w", srcT.Name(), err)
		}
		m.decoders.Store(decName, newTmpl)
		tmpl = newTmpl
	}
	return tmpl.(*StructMapDecoder), nil
}

func (m *Mapper) SetHandle(name string, handle HandleFunc) {
	m.handlers[name] = handle
}
