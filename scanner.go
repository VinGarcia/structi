package structi

import (
	"fmt"
	"reflect"
	"sync"
	"unicode"

	"github.com/vingarcia/structi/internal/types"
	"github.com/vingarcia/structi/tags"
)

// IteratorFunc is the interface that allows the ForEach function to get values
// from any data source and then use these values to fill a targetStruct.
type IteratorFunc func(field Field) error

// Field is the input expected by the `IteratorFunc` and contains all
// the information about the field that is currently being targeted
// by the ForEach() function.
type Field struct {
	*fieldInfo
	Set   func(value any) error
	Value any
}

// fieldInfo contains all the immutable values of
// the Field so that we can keep this info cached.
type fieldInfo struct {
	idx int

	Tags map[string]string
	Name string
	Kind reflect.Kind
	Type reflect.Type

	IsEmbeded bool
}

type StructInfo struct {
	Fields []fieldInfo
}

// GetStructInfo will return (and cache) information about the given struct.
//
// `targetStruct` should either be a pointer to a struct type, or a
// reflect.Type object of the structure in question
func GetStructInfo(targetStruct interface{}) (si StructInfo, err error) {
	if t, ok := targetStruct.(reflect.Type); ok {
		if t.Kind() != reflect.Ptr {
			t = reflect.PointerTo(t)
		}
		_, si.Fields, err = getStructInfoForType(t)
		return si, err
	}

	_, _, si.Fields, err = getStructInfo(targetStruct)
	return si, err
}

// ForEach reads from the input decoder in order to fill the
// attributes of an target struct.
func ForEach(targetStruct interface{}, iterate IteratorFunc) error {
	_, v, fields, err := getStructInfo(targetStruct)
	if err != nil {
		return err
	}

	for _, field := range fields {
		err := iterate(Field{
			fieldInfo: &field,
			Value:     v.Elem().Field(field.idx).Addr().Interface(),
			Set:       setAttrValue(v, field),
		})
		if err != nil {
			return fmt.Errorf("iteration error on field '%s' of type '%v': %w", field.Name, field.Type, err)
		}
	}

	return nil
}

func setAttrValue(structPtrValue reflect.Value, field fieldInfo) func(value any) error {
	return func(value any) error {
		if field.Kind != reflect.Slice {
			convertedValue, err := types.NewConverter(value).Convert(field.Type)
			if err != nil {
				return err
			}

			structPtrValue.Elem().Field(field.idx).Set(convertedValue)
			return nil
		}

		// In case it is a slice

		sliceValue := reflect.ValueOf(value)
		sliceType := sliceValue.Type()
		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem()
			sliceValue = sliceValue.Elem()
		}

		if sliceType.Kind() != reflect.Slice {
			t := structPtrValue.Elem().Type()
			return fmt.Errorf("expected slice for field %#v but got %v of type %v", t.Field(field.idx), sliceValue, sliceType)
		}

		elemType := field.Type.Elem()

		sliceLen := sliceValue.Len()
		targetSlice := reflect.MakeSlice(field.Type, sliceLen, sliceLen)
		for i := 0; i < sliceLen; i++ {
			convertedValue, err := types.NewConverter(sliceValue.Index(i).Interface()).Convert(elemType)
			if err != nil {
				t := structPtrValue.Elem().Type()
				return fmt.Errorf("error converting %v[%d]: %w", t.Field(field.idx).Name, i, err)
			}

			targetSlice.Index(i).Set(convertedValue)
		}

		structPtrValue.Elem().Field(field.idx).Set(targetSlice)

		return nil
	}
}

// This cache is kept as a pkg variable
// because the total number of types on a program
// should be finite. So keeping a single cache here
// works fine.
var structInfoCache = &sync.Map{}

func getStructInfo(targetStruct interface{}) (reflect.Type, reflect.Value, []fieldInfo, error) {
	v := reflect.ValueOf(targetStruct)

	t, fields, err := getStructInfoForType(v.Type())
	if err != nil {
		return nil, reflect.Value{}, nil, err
	}

	// Only validate v after parsing the type, otherwise the call to v.IsNil() might panic.
	if v.IsNil() {
		return nil, reflect.Value{}, nil, fmt.Errorf("expected non-nil pointer to struct, but got: %#v", targetStruct)
	}

	return t, v, fields, err
}

func getStructInfoForType(ptrType reflect.Type) (reflect.Type, []fieldInfo, error) {
	data, found := structInfoCache.Load(ptrType)
	if found {
		return ptrType.Elem(), data.([]fieldInfo), nil
	}

	if ptrType.Kind() != reflect.Ptr {
		return nil, nil, fmt.Errorf("expected struct pointer but got: %v", ptrType)
	}

	t := ptrType.Elem()
	if t.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("can only get struct info from structs, but got: %s", ptrType)
	}

	info := []fieldInfo{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// If it is unexported:
		if unicode.IsLower(rune(field.Name[0])) {
			continue
		}

		parsedTags, err := tags.ParseTags(field.Tag)
		if err != nil {
			return nil, nil, err
		}

		info = append(info, fieldInfo{
			idx:  i,
			Tags: parsedTags,
			Name: field.Name,
			Type: field.Type,
			Kind: field.Type.Kind(),

			// ("Anonymous" is the name for embeded fields on the stdlib)
			IsEmbeded: field.Anonymous,
		})
	}

	structInfoCache.Store(ptrType, info)
	return t, info, nil
}
