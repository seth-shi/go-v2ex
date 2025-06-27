package response

import (
	"strings"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

type V2ReplyResponse struct {
	Result     []V2ReplyResult `json:"result"`
	Pagination PageResponse    `json:"pagination"`
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

	renderContent string
}

func (r *V2ReplyResult) GetContent() string {

	if r.renderContent == "" {
		var content = r.Content
		// 如果是链接出现多次, 那么只保留一次
		list := pkg.ExtractImgURLsNoUnique(content)
		for k, v := range lo.CountValues(list) {
			if v > 1 {
				content = strings.Replace(content, k, "", v-1)
			}
		}

		r.renderContent = pkg.SafeRenderHtml(content)
	}

	return r.renderContent
}
