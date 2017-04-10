package example

// cat ./ranges.go | astdump > ranges.go.ast
func ranges() {
	type uniStruct struct{ value string }
	m := map[string]uniStruct{
		"foo":  uniStruct{"bar"},
		"hoge": uniStruct{"piyo"},
	}
	var indices []*int
	var keys []*string
	var values []*uniStruct
	var functions []func()
	for key, value := range m {
		print(key, value.value) // safe
		func() {
			print(key, value.value) // safe
		}()
		functions = append(functions, func() {
			print(key, value.value) // unsafe
		})
		keys = append(keys, &key)       // unsafe
		values = append(values, &value) // unsafe

		key, value := key, value // safe
		functions = append(functions, func() {
			print(key, value.value) // safe
		})
		keys = append(keys, &key)       // safe
		values = append(values, &value) // safe
	}
	for key := range m {
		print(key) // safe
		func() {
			print(key) // safe
		}()
		functions = append(functions, func() {
			print(key) // unsafe
		})
		keys = append(keys, &key) // unsafe

		key := key // safe
		functions = append(functions, func() {
			print(key) // safe
		})
		keys = append(keys, &key) // safe
	}
	for i, j := 0, 0; i < 1; i++ {
		print(i) // safe
		print(j)
		func() {
			print(i) // safe
			print(j) // safe
		}()
		functions = append(functions, func() {
			print(i) // unsafe
			print(j) // unsafe
		})
		indices = append(indices, &i) // unsafe

		i := i // safe
		j := j // safe
		functions = append(functions, func() {
			print(i) // safe
			print(j) // safe
		})
		indices = append(indices, &i) // safe
		indices = append(indices, &j) // safe
		j++
	}
}

var _ = ranges
