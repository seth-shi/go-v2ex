package types

type V2MemberResponse struct {
	v2ApiError
	Result *V2MemberResult `json:"result"`
}

type V2MemberResult struct {
	Id             int    `json:"id"`
	Username       string `json:"username"`
	Url            string `json:"url"`
	Website        string `json:"website"`
	Twitter        string `json:"twitter"`
	Psn            string `json:"psn"`
	Github         string `json:"github"`
	Btc            string `json:"btc"`
	Location       string `json:"location"`
	Tagline        string `json:"tagline"`
	Bio            string `json:"bio"`
	AvatarMini     string `json:"avatar_mini"`
	AvatarNormal   string `json:"avatar_normal"`
	AvatarLarge    string `json:"avatar_large"`
	AvatarXlarge   string `json:"avatar_xlarge"`
	AvatarXxlarge  string `json:"avatar_xxlarge"`
	AvatarXxxlarge string `json:"avatar_xxxlarge"`
	Created        int    `json:"created"`
	LastModified   int    `json:"last_modified"`
}
