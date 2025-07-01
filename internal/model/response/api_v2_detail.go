package response

import (
	"strings"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

type V2Detail struct {
	Result V2DetailResult `json:"result"`
}

type V2DetailResult struct {
	Id           int64              `json:"id"`
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	Url          string             `json:"url"`
	Replies      int                `json:"replies"`
	LastReplyBy  string             `json:"last_reply_by"`
	Created      int64              `json:"created"`
	LastModified int64              `json:"last_modified"`
	LastTouched  int64              `json:"last_touched"`
	Member       MemberResult       `json:"member"`
	Node         NodeInfoResult     `json:"node"`
	Supplements  []SupplementResult `json:"supplements"`

	renderContent string
}

type SupplementResult struct {
	Id            int    `json:"id"`
	Content       string `json:"content"`
	Created       int64  `json:"created"`
	renderContent string
}

func (r V2DetailResult) GetContent() string {

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

func (r SupplementResult) GetContent() string {
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
