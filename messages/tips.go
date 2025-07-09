package messages

type ShowStatusBarTextRequest struct {
	FirstText string
	HelpText  string
}

type ProxyShowToastRequest struct {
	Text string
}
