package custom

func bodies(x int, ch chan int) {
	switch x {
	case 1: // want `case clause body spans 4 lines, max is 3`
		x++
		x++
		x++
		x++
	case 2: // exactly 3 lines: not reported
		x++
		x++
		x++
	case 3: // empty body: not reported
	case 4: // want `case clause body spans 4 lines, max is 3`
		x++
		// interior comment lines count towards the span
		x++
		x++
	case 5: // leading and trailing comments do not count
		// leading comment
		x++
		x++
		x++
		// trailing comment
	case 7: // want `case clause body spans 4 lines, max is 3`
		x++

		x++
		x++
	case 6: // want `case clause body spans 4 lines, max is 3`
		println(
			x,
			x,
		)
	default: // want `case clause body spans 4 lines, max is 3`
		x++
		x++
		x++
		x++
	}

	select {
	case <-ch: // want `case clause body spans 4 lines, max is 3`
		x++
		x++
		x++
		x++
	case ch <- x: // 3 lines: not reported
		x++
		x++
		x++
	}
}

func typeSwitch(v any) {
	switch v.(type) {
	case int: // want `case clause body spans 4 lines, max is 3`
		println(1)
		println(2)
		println(3)
		println(4)
	case string:
		println(1)
	}
}
