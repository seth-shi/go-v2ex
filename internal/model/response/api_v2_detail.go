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
	Id              int64  `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	ContentRendered string `json:"content_rendered"`
	Syntax          int    `json:"syntax"`
	Url             string `json:"url"`
	Replies         int    `json:"replies"`
	LastReplyBy     string `json:"last_reply_by"`
	Created         int64  `json:"created"`
	LastModified    int64  `json:"last_modified"`
	LastTouched     int64  `json:"last_touched"`
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
	Node struct {
		Id           int    `json:"id"`
		Url          string `json:"url"`
		Name         string `json:"name"`
		Title        string `json:"title"`
		Header       string `json:"header"`
		Footer       string `json:"footer"`
		Avatar       string `json:"avatar"`
		Topics       int    `json:"topics"`
		Created      int64  `json:"created"`
		LastModified int64  `json:"last_modified"`
	} `json:"node"`
	Supplements []SupplementResult `json:"supplements"`

	renderContent string
}

type SupplementResult struct {
	Id              int    `json:"id"`
	Content         string `json:"content"`
	ContentRendered string `json:"content_rendered"`
	Syntax          int    `json:"syntax"`
	Created         int64  `json:"created"`

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
