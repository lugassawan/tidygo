package example

// Reported: switch inside a switch case.
func switchInSwitch(a, b int) {
	switch a {
	case 1:
		switch b { // want `switch statement nested inside another switch`
		case 2:
		}
	}
}

// Reported: type switch inside a switch.
func typeSwitchInSwitch(a int, v any) {
	switch a {
	case 1:
		switch v.(type) { // want `switch statement nested inside another switch`
		case int:
		}
	}
}

// Reported: switch inside a type switch.
func switchInTypeSwitch(v any, b int) {
	switch v.(type) {
	case int:
		switch b { // want `switch statement nested inside another switch`
		case 1:
		}
	}
}

// Reported: switch inside a default clause.
func switchInDefault(a, b int) {
	switch a {
	default:
		switch b { // want `switch statement nested inside another switch`
		case 1:
		}
	}
}

// Reported: if/for blocks between the switches do not stop the walk.
func switchInIfFor(a, b int) {
	switch a {
	case 1:
		if b > 0 {
			for i := 0; i < b; i++ {
				switch i { // want `switch statement nested inside another switch`
				case 1:
				}
			}
		}
	}
}

// Reported twice: three levels deep.
func threeLevels(a, b, c int) {
	switch a {
	case 1:
		switch b { // want `switch statement nested inside another switch`
		case 1:
			switch c { // want `switch statement nested inside another switch`
			case 1:
			}
		}
	}
}

// Not reported: a function literal starts a new scope.
func inFuncLit(a, b int) {
	switch a {
	case 1:
		f := func() {
			switch b {
			case 1:
			}
		}
		f()
	}
}

// Reported: nesting inside a function literal is still caught within the closure.
func nestedInFuncLit(a, b, c int) {
	switch a {
	case 1:
		f := func() {
			switch b {
			case 1:
				switch c { // want `switch statement nested inside another switch`
				case 1:
				}
			}
		}
		f()
	}
}

// Not reported: the nearest switch-like statement is a select.
func inSelect(a int, ch chan int) {
	switch a {
	case 1:
		select {
		case <-ch:
			switch a {
			case 1:
			}
		}
	}
}

// Not reported: sibling switches.
func siblings(a, b int) {
	switch a {
	case 1:
	}
	switch b {
	case 1:
	}
}
