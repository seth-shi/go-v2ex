package messages

type ShowAlertRequest struct {
	Text string
	Help string
}

type ShowToastRequest struct {
	Text string
}

type ShiftToastRequest struct{}
