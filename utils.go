package jsonmanu

import (
	"fmt"
	"reflect"
)

func isMap(t any) bool {
	return reflect.ValueOf(t).Kind() == reflect.Map
}

func isSlice(t any) bool {
	return reflect.ValueOf(t).Kind() == reflect.Slice
}

func isMapOrSlice(t any) bool {
	return isMap(t) || isSlice(t)
}

func flattenArray(arr []any) (result []any) {
	for _, item := range arr {
		if !isSlice(item) {
			result = append(result, item)
			continue
		}

		flattenedItem := flattenArray(item.([]any))

		for _, fitem := range flattenedItem {
			result = append(result, fitem)
		}
	}

	return
}

func mapGetDeep(m any, key string) (result []any) {
	if isMap(m) {
		v, ok := m.(map[string]any)[key]
		if ok {
			result = append(result, v)
			return
		}

		for mkey := range iterMapKeys(m, nil) {
			v, _ := m.(map[string]any)[mkey]
			if isMapOrSlice(v) {
				if dv := mapGetDeep(v, key); dv != nil {
					result = append(result, dv)
				}
			}
		}
	}

	if isSlice(m) {
		for _, item := range m.([]any) {
			if v := mapGetDeep(item, key); v != nil {
				result = append(result, v)
			}
		}
	}

	return
}

func mapPutDeep(m any, key string, value any) error {
	if isMap(m) {
		if _, ok := m.(map[string]any)[key]; ok {
			m.(map[string]any)[key] = value
			return nil
		}

		for mkey := range iterMapKeys(m, nil) {
			mvalue, _ := m.(map[string]any)[mkey]
			if isMapOrSlice(mvalue) {
				mapPutDeep(mvalue, key, value)
			}
		}
	}

	if isSlice(m) {
		for _, item := range m.([]any) {
			mapPutDeep(item, key, value)
		}
	}
	return nil
}

func mapGetDeepFlattened(m any, key string) []any {
	return flattenArray(mapGetDeep(m, key))
}

type ArrayX map[any]int

func arrayxFromArray(arr []any) ArrayX {
	result := make(ArrayX)
	for _, item := range arr {
		itemx, ok := result[item]
		if !ok {
			result[item] = 1
		} else {
			result[item] = itemx + 1
		}
	}

	return result
}

func compareArrayX(arrx1 ArrayX, arrx2 ArrayX) bool {
	if len(arrx1) != len(arrx2) {
		return false
	}

	for k, v := range arrx1 {
		if arrx2[k] != v {
			return false
		}
	}
	return true
}

func sendOrQuit[T any](t T, out chan<- T, quit <-chan struct{}) bool {
	select {
	case out <- t:
		return true
	case <-quit:
		return false
	}
}

func iterMapKeys(m any, quit <-chan struct{}) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)

		vm := reflect.ValueOf(m)
		if vm.Kind() == reflect.Map {
			for _, key := range vm.MapKeys() {
				skey := fmt.Sprintf("%v", key)
				sendOrQuit(skey, out, quit)
			}
		}
	}()
	return out
}

func mapHasKey(m any, key string) bool {
	for mkey := range iterMapKeys(m, nil) {
		if mkey == key {
			return true
		}
	}
	return false
}
