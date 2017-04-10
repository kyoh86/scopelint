package example

import "fmt"

func readme() {
	values := []string{"a", "b", "c"}
	var funcs []func()
	for _, val := range values {
		funcs = append(funcs, func() {
			fmt.Println(val)
		})
	}
	for _, f := range funcs {
		f()
	}
	/*output:
	  c
	  c
	  c
	  (unstable)*/
	var copies []*string
	for _, val := range values {
		copies = append(copies, &val)
	}
	/*(in copies)
	  &"c"
	  &"c"
	  &"c"
	*/
}
