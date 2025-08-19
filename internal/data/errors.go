package data

type LoadApiDataErr struct {
	Err error
	Msg string
}

func (e LoadApiDataErr) Error() string {
	return e.Err.Error()
}
