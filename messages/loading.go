package messages

type StartLoading struct {
	Text string
	ID   int
}

type EndLoading struct {
	ID int
}
