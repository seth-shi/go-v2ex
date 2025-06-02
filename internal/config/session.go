package config

var (
	Session = &sessionData{}
)

// 跳转的时候模型和 routes.x 不是同一个, 被复制了, 所以存储到全局中
type sessionData struct {
	TopicPage        int
	TopicActiveIndex int
}
