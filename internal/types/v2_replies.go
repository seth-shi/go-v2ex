package types

import "jaytaylor.com/html2text"

type V2ReplyResponse struct {
	V2ApiError
	Result []V2ReplyResult `json:"result"`
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
	if r.ContentRendered != "" {
		if text, err := html2text.FromString(r.ContentRendered, html2text.Options{PrettyTables: true}); err == nil {
			return text
		}
	}

	return r.Content
}
