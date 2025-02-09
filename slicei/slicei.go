package slicei

import (
	"fmt"
	"reflect"

	"github.com/vingarcia/structi/internal/types"
)

// IteratorFunc is the interface that allows the ForEach function to get values
// from any data source and then use these values to fill a targetStruct.
type IteratorFunc func(field Field) error

// Field is the input expected by the `IteratorFunc` and contains all
// the information about the field that is currently being targeted
// by the ForEach() function.
type Field struct {
	Index int
	Kind  reflect.Kind
	Type  reflect.Type
	Value any

	Set func(value any) error
}

func Append(targetSlice any, items ...any) error {
	t, v, err := getSliceInfo(targetSlice)
	if err != nil {
		return err
	}

	elemType := t.Elem()
	sliceValue := v.Elem()
	for _, item := range items {
		convertedValue, err := types.NewConverter(item).Convert(elemType)
		if err != nil {
			return fmt.Errorf("error converting %+v to %v: %w", item, elemType, err)
		}

		sliceValue = reflect.Append(sliceValue, convertedValue)
	}

	v.Elem().Set(sliceValue)

	return nil
}

// ForEach iterates over the slice calling the iterate function
// for each item
func ForEach(targetSlice interface{}, iterate IteratorFunc) error {
	_, v, err := getSliceInfo(targetSlice)
	if err != nil {
		return err
	}

	sliceLen := v.Elem().Len()
	for i := 0; i < sliceLen; i++ {
		t := v.Elem().Index(i).Type()
		err := iterate(Field{
			Index: i,
			Kind:  t.Kind(),
			Type:  t,
			Value: v.Elem().Index(i).Addr().Interface(),

			Set: setItemValue(v.Elem(), t, i),
		})
		if err != nil {
			return fmt.Errorf("iteration error on item '%d' of type '%v': %w", i, t, err)
		}
	}

	return nil
}

func setItemValue(sliceValue reflect.Value, itemType reflect.Type, index int) func(value any) error {
	return func(value any) error {
		convertedValue, err := types.NewConverter(value).Convert(itemType)
		if err != nil {
			return fmt.Errorf("error converting %v[%d]: %w", sliceValue.Type(), index, err)
		}

		sliceValue.Index(index).Set(convertedValue)
		return nil
	}
}

func getSliceInfo(targetSlice interface{}) (reflect.Type, reflect.Value, error) {
	if targetSlice == nil {
		return nil, reflect.Value{}, fmt.Errorf("unexpected nil input")
	}

	v, ok := targetSlice.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(targetSlice)
	}
	ptrType := v.Type()

	if ptrType.Kind() != reflect.Ptr {
		return nil, reflect.Value{}, fmt.Errorf("expected slice pointer but got: %v", ptrType)
	}

	t := ptrType.Elem()
	if t.Kind() != reflect.Slice {
		return nil, reflect.Value{}, fmt.Errorf("can only get slice info from slices, but got: %s", ptrType)
	}

	return t, v, nil
}
