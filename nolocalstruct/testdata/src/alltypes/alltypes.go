package alltypes

func locals() {
	type Iface interface{ Do() }  // want `named type "Iface" should be declared at package level, not inside a function`
	type ID int                   // want `named type "ID" should be declared at package level, not inside a function`
	type Handler func()           // want `named type "Handler" should be declared at package level, not inside a function`
	type Box[T any] struct{ v T } // want `named struct type "Box" should be declared at package level, not inside a function`
	type List[T any] []T          // want `named type "List" should be declared at package level, not inside a function`
	type Plain struct{ x int }    // want `named struct type "Plain" should be declared at package level, not inside a function`

	// Aliases are not new named types, so they are skipped...
	type Alias = int
	type IfaceAlias = interface{ Do() }
	// ...unless they alias a struct, which the default rule already reports.
	type StructAlias = struct{ y int } // want `named struct type "StructAlias" should be declared at package level, not inside a function`

	_ = func() {
		type InLiteral int // want `named type "InLiteral" should be declared at package level, not inside a function`
	}
}
