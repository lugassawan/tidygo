package example

func bodies(x int) {
	switch x {
	case 1: // want `case clause body spans 11 lines, max is 10`
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
	case 2:
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
		x++
	}
}
