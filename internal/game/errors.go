package game

type ErrWrongTurn struct {
	Msg string
}

func (e ErrWrongTurn) Error() string { return e.Msg }

type ErrMatchNotInProgress struct {
	Msg string
}

func (e ErrMatchNotInProgress) Error() string { return e.Msg }

type ErrIllegalMove struct {
	Msg string
}

func (e ErrIllegalMove) Error() string { return e.Msg }
