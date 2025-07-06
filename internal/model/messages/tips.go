package messages

type ShowStatusBarTextRequest struct {
	FirstText  string
	SecondText string
}

type ProxyShowToastRequest struct {
	Text string
}
