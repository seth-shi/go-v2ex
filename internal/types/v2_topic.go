package types

import (
	"encoding/json"
)

type TopicResource struct {
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

func GetDefaultTopic() *TopicResource {
	jsonText := `  {
        "id" : 128177,
        "title" : "vim\u002Dtranslator1111111111111111111111111111",
        "url" : "http://www.v2ex.com/t/128177",
        "content" : "一个轻巧的vim下的翻译和发音插件，依赖于 google\u002Dtranslator\u002Dcli，或者其他的命令行翻译，查询工具。发音取自google...\u000D\u000A\u000D\u000Ahttps://github.com/farseer90718/vim\u002Dtranslator\u000D\u000A\u000D\u000A功能比较简单。暂时只是实现了个人的需求...",
        "content_rendered" : "一个轻巧的vim下的翻译和发音插件，依赖于 google\u002Dtranslator\u002Dcli，或者其他的命令行翻译，查询工具。发音取自google...\u003Cbr /\u003E\u003Cbr /\u003E\u003Ca target\u003D\u0022_blank\u0022 href\u003D\u0022https://github.com/farseer90718/vim\u002Dtranslator\u0022 rel\u003D\u0022nofollow\u0022\u003Ehttps://github.com/farseer90718/vim\u002Dtranslator\u003C/a\u003E\u003Cbr /\u003E\u003Cbr /\u003E功能比较简单。暂时只是实现了个人的需求...",
        "replies" : 0,
        "member" : {
            "id" : 67060,
            "username" : "farseer2014",
            "tagline" : "",
            "avatar_mini" : "//cdn.v2ex.com/avatar/6766/2b3d/67060_mini.png?m=1408121347",
            "avatar_normal" : "//cdn.v2ex.com/avatar/6766/2b3d/67060_normal.png?m=1408121347",
            "avatar_large" : "//cdn.v2ex.com/avatar/6766/2b3d/67060_large.png?m=1408121347"
        },
        "node" : {
            "id" : 17,
            "name" : "create",
            "title" : "分享创造",
            "title_alternative" : "Create",
            "url" : "http://www.v2ex.com/go/create",
            "topics" : 2621,
            "avatar_mini" : "//cdn.v2ex.com/navatar/70ef/df2e/17_mini.png?m=1388448923",
            "avatar_normal" : "//cdn.v2ex.com/navatar/70ef/df2e/17_normal.png?m=1388448923",
            "avatar_large" : "//cdn.v2ex.com/navatar/70ef/df2e/17_large.png?m=1388448923"
        },
        "created" : 1408122614,
        "last_modified" : 1408122614,
        "last_touched" : 1408122434
    }`
	var tt TopicResource
	json.Unmarshal([]byte(jsonText), &tt)
	return &tt
}
