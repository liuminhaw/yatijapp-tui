package data

type LoadApiDataErr struct {
	Status int
	Err    error
	Msg    string
}

func (e LoadApiDataErr) Error() string {
	return e.Err.Error()
}
