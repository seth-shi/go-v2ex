package response

import (
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

type V2Reply struct {
	Result     []V2ReplyResult `json:"result"`
	Pagination Page            `json:"pagination"`
}

type V2ReplyResult struct {
	Id              int    `json:"id"`
	Content         string `json:"content"`
	ContentRendered string `json:"content_rendered"`
	Created         int64  `json:"created"`
	Member          struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
		Bio      string `json:"bio"`
		Website  string `json:"website"`
		Github   string `json:"github"`
		Url      string `json:"url"`
		Avatar   string `json:"avatar"`
		Created  int64  `json:"created"`
	} `json:"member"`
}

func (r V2ReplyResult) GetContent() string {
	var content = r.ContentRendered
	if r.ContentRendered == "" {
		content = r.Content
	}

	return pkg.SafeRenderHtml(content)
}
