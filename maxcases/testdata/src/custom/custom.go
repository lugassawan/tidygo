package custom

// Reported: 4 clauses with default counted exceeds the configured max of 3.
func switchOver(x int) {
	switch x { // want `switch has 4 case clauses, max is 3`
	case 1:
	case 2:
	case 3:
	default:
	}
}

// Not reported: exactly 3 clauses.
func switchAtLimit(x int) {
	switch x {
	case 1:
	case 2:
	default:
	}
}

// Reported: type switches are counted too.
func typeSwitchOver(v any) {
	switch v.(type) { // want `type switch has 4 case clauses, max is 3`
	case int:
	case string:
	case bool:
	default:
	}
}

// Not reported: 3 type switch clauses.
func typeSwitchAtLimit(v any) {
	switch v.(type) {
	case int:
	case string:
	default:
	}
}

// Reported: select statements are counted too.
func selectOver(a, b, c chan int) {
	select { // want `select has 4 case clauses, max is 3`
	case <-a:
	case <-b:
	case <-c:
	default:
	}
}

// Not reported: 3 select clauses.
func selectAtLimit(a, b chan int) {
	select {
	case <-a:
	case <-b:
	default:
	}
}
