package struct_mapper

import (
	"errors"
	"fmt"
	"github.com/gvassili/tag_parser"
	"reflect"
)

type fieldMapNode struct {
	children []fieldMapChild
	fields   []FieldDef
}

type fieldMapChild struct {
	dstIdx int
	node   *fieldMapNode
}

type FieldDef struct {
	srcIdx int
	dstIdx int
	tag    tag
}

type structMapDecoderChild struct {
	structIdx    int
	fieldMapRoot fieldMapNode
	children     []structMapDecoderChild
}

type tag struct {
	Path   []string
	Handle HandleFunc
}

var ignoreTag = errors.New("ignore tag")

func parseTag(s string, handlers map[string]HandleFunc) (tag, error) {
	var tag tag
	params, err := tag_parser.Parse(s)
	if err != nil {
		return tag, fmt.Errorf("\"%s\": %w", s, err)
	}
	if len(params) > 2 {
		return tag, errors.New("more than 2 tag parameter")
	}
	if len(params) == 0 {
		return tag, ignoreTag
	}
	for i, p := range params {
		if i == 0 {
			if len(p.Values) == 0 {
				tag.Path = []string{p.Key}
			} else {
				switch p.Key {
				case "-":
					return tag, ignoreTag
				case "path":
					if len(p.Values) == 0 {
						return tag, errors.New("path tag parameter should not be empty")
					}
					tag.Path = p.Values
				default:
					return tag, fmt.Errorf("unknow parameter \"%s\"", p.Key)
				}
			}
		} else if i == 1 {
			if len(p.Values) > 0 {
				return tag, errors.New("second parameter shouldn't have value")
			}
			h, ok := handlers[p.Key]
			if !ok {
				return tag, fmt.Errorf("handle %s doesn't exist", p.Key)
			}
			tag.Handle = h
		}
	}
	return tag, nil
}

func newStructDecoderChild(srcT reflect.Type, dstT reflect.Type, handles map[string]HandleFunc) (structMapDecoderChild, error) {
	rw := structMapDecoderChild{}
	nodeMap := make(map[string]*fieldMapNode)
	if srcT.Kind() == reflect.Ptr {
		srcT = srcT.Elem()
	}
	nf := srcT.NumField()
	for i := 0; i < nf; i++ {
		f := srcT.Field(i)
		if f.Type.Kind() == reflect.Struct || f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct {
			child, err := newStructDecoderChild(f.Type, dstT, handles)
			if err != nil {
				return rw, fmt.Errorf("instanciate child row writer %s: %w", f.Name, err)
			} else if child.hasField() {
				child.structIdx = i
				rw.children = append(rw.children, child)
			}
		}
		tag, err := parseTag(f.Tag.Get("map"), handles)
		if err == ignoreTag {
			continue
		} else if err != nil {
			return rw, fmt.Errorf("parse tag: %w", err)
		}
		var getDstField func(dstT reflect.Type, path []string, idxs []int) ([]int, reflect.Type, error)
		getDstField = func(dstT reflect.Type, path []string, idxs []int) ([]int, reflect.Type, error) {
			if dstT.Kind() == reflect.Ptr {
				dstT = dstT.Elem()
			}
			if len(path) > 0 {
				f, ok := dstT.FieldByName(path[0])
				if !ok {
					return nil, nil, fmt.Errorf("field %s doesn't exist in type %s", path[0], dstT.Name())
				}
				idxs = append(idxs, f.Index[0])
				dstT = f.Type
			}
			var err error
			if len(path) > 1 {
				idxs, dstT, err = getDstField(dstT, path[1:], idxs)
			}
			return idxs, dstT, err
		}

		idxs, t, err := getDstField(dstT, tag.Path, nil)
		if err != nil {
			return rw, fmt.Errorf("could not decode destination path: %w", err)
		}

		if tag.Handle == nil && t.Name() != f.Type.Name() && (f.Type.Kind() == reflect.Ptr && t.Name() != f.Type.Elem().Name()) {
			return rw, fmt.Errorf("can not handle type %s with type %s of field %s", t.Name(), f.Type.Name(), f.Name)
		}

		var insertField func(parent *fieldMapNode, i int)
		insertField = func(parent *fieldMapNode, j int) {
			if j == len(idxs)-1 {
				parent.fields = append(parent.fields, FieldDef{
					srcIdx: i,
					dstIdx: idxs[j],
					tag:    tag,
				})
			} else {
				key := fmt.Sprint(idxs[:j+1])
				child, ok := nodeMap[key]
				if !ok {
					child = new(fieldMapNode)
					nodeMap[key] = child
					parent.children = append(parent.children, fieldMapChild{
						dstIdx: idxs[j],
						node:   child,
					})
				}
				insertField(child, j+1)
			}
		}
		insertField(&rw.fieldMapRoot, 0)
	}
	return rw, nil
}

func (mc *structMapDecoderChild) hasField() bool {
	var checkFields func(node fieldMapNode) bool
	checkFields = func(node fieldMapNode) bool {
		if len(node.fields) > 0 {
			return true
		}
		for _, c := range node.children {
			if checkFields(*c.node) {
				return true
			}
		}
		return false
	}
	if checkFields(mc.fieldMapRoot) {
		return true
	}
	for _, c := range mc.children {
		if c.hasField() {
			return true
		}
	}
	return false
}

func (mc *structMapDecoderChild) decode(src reflect.Value, dst reflect.Value) error {
	var writeField func(node fieldMapNode, v reflect.Value)
	writeField = func(node fieldMapNode, v reflect.Value) {
		for _, c := range node.children {
			fv := v.Field(c.dstIdx)
			if fv.Type().Kind() == reflect.Ptr {
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				fv = fv.Elem()
			}
			writeField(*c.node, fv)
		}
		for _, f := range node.fields {
			fv := v.Field(f.dstIdx)
			if fv.Type().Kind() == reflect.Ptr {
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				fv = fv.Elem()
			}
			srcV := src.Field(f.srcIdx)
			if srcV.Kind() == reflect.Ptr {
				if !srcV.IsNil() {
					srcV = srcV.Elem()
				} else {
					continue
				}
			}
			if f.tag.Handle != nil {
				fv.Set(reflect.ValueOf(f.tag.Handle(srcV.Interface())))
			} else {
				fv.Set(srcV)
			}
		}
	}
	writeField(mc.fieldMapRoot, dst)
	for _, ct := range mc.children {
		fv := src.Field(ct.structIdx)
		if fv.Type().Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			fv = fv.Elem()
		}
		if err := ct.decode(fv, dst); err != nil {
			return err
		}
	}
	return nil
}

type StructMapDecoder struct {
	root   structMapDecoderChild
	handle map[string]HandleFunc
}

func newStructMapDecoder(srcT reflect.Type, dstT reflect.Type, handle map[string]HandleFunc) (*StructMapDecoder, error) {
	rw := &StructMapDecoder{}
	root, err := newStructDecoderChild(srcT, dstT, handle)
	if err != nil {
		return nil, fmt.Errorf("instanciating child struct map decoder %s: %w", srcT.Name(), err)
	}
	rw.root = root
	return rw, nil
}

func (m *StructMapDecoder) Decode(src interface{}, dst interface{}) error {
	srcR := reflect.ValueOf(src)
	if srcR.Type().Kind() != reflect.Ptr || reflect.Indirect(srcR).Type().Kind() != reflect.Struct {
		panic(errors.New("src interface must be pointer"))
	}
	srcV := reflect.Indirect(srcR)
	dstR := reflect.ValueOf(dst)
	if dstR.Type().Kind() != reflect.Ptr || reflect.Indirect(dstR).Type().Kind() != reflect.Struct {
		panic(errors.New("dst interface must be pointer"))
	}
	dstV := reflect.Indirect(dstR)
	return m.root.decode(srcV, dstV)
}
