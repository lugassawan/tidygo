package splitfile

func (t T) First() {} // OK - a.go's unexported declarations belong to another file

func (t T) helper() {} // OK

func (t T) Shown() {} // want `exported method T.Shown should appear before unexported methods of T`
