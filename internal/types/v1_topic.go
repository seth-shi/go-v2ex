package types

type V1TopicResult struct {
	Id              int    `json:"id"`
	Title           string `json:"title"`
	Url             string `json:"url"`
	Content         string `json:"content"`
	ContentRendered string `json:"content_rendered"`
	Replies         int    `json:"replies"`
	Member          struct {
		Id           int    `json:"id"`
		Username     string `json:"username"`
		Tagline      string `json:"tagline"`
		AvatarMini   string `json:"avatar_mini"`
		AvatarNormal string `json:"avatar_normal"`
		AvatarLarge  string `json:"avatar_large"`
	} `json:"member"`
	Node struct {
		Id               int    `json:"id"`
		Name             string `json:"name"`
		Title            string `json:"title"`
		TitleAlternative string `json:"title_alternative"`
		Url              string `json:"url"`
		Topics           int    `json:"topics"`
		AvatarMini       string `json:"avatar_mini"`
		AvatarNormal     string `json:"avatar_normal"`
		AvatarLarge      string `json:"avatar_large"`
	} `json:"node"`
	Created      int64 `json:"created"`
	LastModified int64 `json:"last_modified"`
	LastTouched  int64 `json:"last_touched"`
}
