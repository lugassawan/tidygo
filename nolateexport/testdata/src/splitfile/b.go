package splitfile

func (t T) helper() {} // OK

func (t T) Shown() {} // want `exported method T.Shown should appear before unexported methods of T`
