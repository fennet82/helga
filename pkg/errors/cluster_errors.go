package errors

type ErrHelmClient struct {
	ErrMsg string
}

func (e ErrHelmClient) Error() string {
	return e.ErrMsg
}

type ErrInSyncProcess struct {
	ErrMsg string
}

func (e ErrInSyncProcess) Error() string {
	return e.ErrMsg
}
