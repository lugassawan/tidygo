package suffix

type Order struct{ total int }

func (o Order) Total() int { return o.total }
