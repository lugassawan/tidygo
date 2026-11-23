package example

// Every pair below differs in one detail only, so none of them is reported.
// Each body has three statements, the smallest size that is compared, and the
// signatures are identical, so only the body tells them apart.

var pk int

func sink(...any) {}

func id(v ...int) int { return len(v) }

func opAdd(a, b int) int {
	x := a + b
	sink(x)
	return x
}

func opSub(a, b int) int {
	x := a - b
	sink(x)
	return x
}

func unaryNeg(a int) int {
	x := -a
	sink(x)
	return x
}

func unaryPlus(a int) int {
	x := +a
	sink(x)
	return x
}

func incr(a int) int {
	a++
	sink(a)
	return a
}

func decr(a int) int {
	a--
	sink(a)
	return a
}

func rangeDefine(items []int) {
	for pk := range items {
		sink(pk)
	}
	sink(items)
	sink(items)
}

func rangeAssign(items []int) {
	for pk = range items {
		sink(pk)
	}
	sink(items)
	sink(items)
}

func declVar() int {
	var n = 1
	sink(n)
	return n
}

func declConst() int {
	const n = 1
	sink(n)
	return n
}

func spread(xs []any) {
	sink(xs...)
	sink(xs...)
	sink(xs...)
}

func noSpread(xs []any) {
	sink(xs)
	sink(xs)
	sink(xs)
}

func aliasDecl() {
	type T = int
	var t T
	sink(t)
}

func defDecl() {
	type T int
	var t T
	sink(t)
}

func chanSend() {
	var c chan<- int
	sink(c)
	sink(c)
}

func chanRecv() {
	var c <-chan int
	sink(c)
	sink(c)
}

func brk(items []string) {
	for _, s := range items {
		if len(s) == 0 {
			break
		}
		sink(s)
	}
	sink(items)
	sink(items)
}

func cont(items []string) {
	for _, s := range items {
		if len(s) == 0 {
			continue
		}
		sink(s)
	}
	sink(items)
	sink(items)
}

// The call tree differs: id(a) and b are two arguments, id(a, b) is one.
func nestedA(a, b int) {
	sink(id(a), b)
	sink(a)
	sink(b)
}

func nestedB(a, b int) {
	sink(id(a, b))
	sink(a)
	sink(b)
}

// The same children walk differently split across the sides of the assignment.
func threeFromOne(f func() (int, int, int), g func() int) {
	var a, b, c int
	a, b, c = f()
	sink(a, b, c)
}

func twoFromTwo(g func() (int, int, int), f func() int) {
	var a, b, c int
	a, b = c, f()
	sink(a, b, c)
}

// A for clause holding only an init statement differs from one holding only a post statement.
func forInit(i int) {
	for i++; ; {
		break
	}
	sink(i)
	sink(i)
}

func forPost(i int) {
	for ; ; i++ {
		break
	}
	sink(i)
	sink(i)
}

func commentedLocal(x int) int {
	// doubled value
	var y = x * 2
	sink(y)
	return y
}

// Reported: a comment on a local declaration does not make the bodies different.
func plainLocal(x int) int { // want `function plainLocal has the same body as commentedLocal \(e\.go:178\)`
	var y = x * 2
	sink(y)
	return y
}

// Not reported: several init functions may repeat each other.
func init() {
	sink(1)
	sink(2)
	sink(3)
}

func init() {
	sink(1)
	sink(2)
	sink(3)
}

// A comm clause with a communication differs from a default clause whose body starts with the same receive.
func selectCase(a chan int) {
	sink(1)
	sink(2)
	select {
	case <-a:
		sink(3)
	}
}

func selectDefault(a chan int) {
	sink(1)
	sink(2)
	select {
	default:
		<-a
		sink(3)
	}
}

type T []int

// A typed literal differs from an elided one whose first element is that type's name.
func litTyped(n int) {
	sink(1)
	sink(2)
	sink([][]int{T{1, 2}})
}

func litElided(T int) {
	sink(1)
	sink(2)
	sink([][]int{{T, 1, 2}})
}
