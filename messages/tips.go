package messages

type ShowStatusBarTextRequest struct {
	Page      string
	FirstText string
	HelpText  string
}

type ProxyShowToastRequest struct {
	Text string
}
