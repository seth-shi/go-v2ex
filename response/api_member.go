package response

import (
	"fmt"
	"strings"

	"github.com/seth-shi/go-v2ex/v2/styles"
)

type MemberResult struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Pro      int    `json:"pro"`
}

func (m MemberResult) GetUserNameLabel(meId int) string {
	var labels []string
	if m.Pro > 0 {
		labels = append(labels, styles.MemberPro)
	}

	if m.Id == meId {
		labels = append(labels, styles.MemberMe)
	}

	return fmt.Sprintf("%s%s", m.Username, strings.Join(labels, ""))
}
