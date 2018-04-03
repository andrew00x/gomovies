package util

func Contains(arr []string, el string) bool {
	for _, cur := range arr {
		if el == cur {
			return true
		}
	}
	return false
}

/*
func ContainsReflect(arr interface{}, predicate interface{}) bool {
	var predicateValue reflect.Value
	if predicateType := reflect.TypeOf(predicate); reflect.TypeOf(predicate).Kind() == reflect.Func && predicateType.NumIn() == 1 && predicateType.NumOut() == 1 && predicateType.Out(0).Kind() == reflect.Bool {
		predicateValue = reflect.ValueOf(predicate)
	} else {
		panic("second argument must be a function that consumes single argument and returns boolean")
	}
	if arrKind := reflect.TypeOf(arr).Kind(); arrKind == reflect.Array || arrKind == reflect.Slice {
		arrValue := reflect.ValueOf(arr)
		for i := 0; i < arrValue.Len(); i++ {
			cur := arrValue.Index(i)
			if predicateValue.Call([]reflect.Value{cur})[0].Bool() {
				return true
			}
		}
		return false
	}
	panic("first argument must be an either array or slice")
}
*/