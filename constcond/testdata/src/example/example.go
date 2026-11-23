package example

type config struct{ Debug bool }

func f() bool { return true }

func conditions(cfg config, ok bool) {
	if true { // want `condition is always true; remove the branch or use a real condition`
		_ = cfg
	}

	if false { // want `condition is always false; remove the branch or use a real condition`
		_ = cfg
	}

	if ok {
		_ = cfg
	} else if true { // want `condition is always true; remove the branch or use a real condition`
		_ = cfg
	}

	if x := f(); true { // want `condition is always true; remove the branch or use a real condition`
		_ = x
	}

	for false { // want `loop condition is always false; remove the loop or use a real condition`
		_ = cfg
	}

	for true { // want `loop condition is always true; remove the loop or use a real condition`
		_ = cfg
	}

	for i := 0; true; i++ { // want `loop condition is always true; remove the loop or use a real condition`
		_ = i
	}
}

func realConditions(cfg config) {
	if cfg.Debug {
		_ = cfg
	}

	if x := f(); x {
		_ = x
	}

	for {
		break
	}

	for i := 0; i < 3; i++ {
		_ = i
	}
}
