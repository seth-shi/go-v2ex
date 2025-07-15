package g

import (
	"strings"

	"github.com/samber/lo"
)

const (
	NodesMy      = "my_nodes"
	LatestNode   = "latest"
	HotNode      = "hot"
	myNodesTitle = "分享发现 · 分享创造 · 问与答 · 酷工作 · 程序员 · 职场话题 · 投资 · 奇思妙想 · 硬件 · 游戏开发"
)

var (
	myNodesKey = []string{
		"share",
		"create",
		"qna",
		"jobs",
		"programmer",
		"career",
		"invest",
		"ideas",
		"hardware",
		"gamedev",
	}
	OfficialNodes = []GroupNode{
		{
			Key:  "tech",
			Name: "技术",
			Nodes: []string{
				"programmer",
				"python",
				"idev",
				"android",
				"linux",
				"nodejs",
				"cloud",
				"bb",
			},
			NodesTitle: "程序员 · Python · iDev · Android · Linux · Node.js · 云计算 · 宽带症候群",
		},
		{
			Key:  "creative",
			Name: "创意",
			Nodes: []string{
				"create",
				"design",
				"ideas",
			},
			NodesTitle: "分享创造 · 设计 · 奇思妙想",
		},
		{
			Key:  "play",
			Name: "好玩",
			Nodes: []string{
				"share",
				"games",
				"movie",
				"tv",
				"music",
				"travel",
				"android",
				"afterdark",
			},
			NodesTitle: "分享发现 · 游戏 · 电影 · 剧集 · 音乐 · 旅行 · Android · 天黑以后",
		},
		{
			Key:  "apple",
			Name: "Apple",
			Nodes: []string{
				"macos",
				"iphone",
				"ipad",
				"macmini",
				"mbp",
				"imac",
				"watch",
				"apple",
			},
			NodesTitle: "macOS · iPhone · iPad · Mac mini · MacBook Pro · iMac ·  WATCH · Apple",
		},
		{
			Key:  "jobs",
			Name: "酷工作",
			Nodes: []string{
				"jobs",
				"cv",
				"career",
				"meet",
				"outsourcing",
			},
			NodesTitle: "酷工作 · 求职 · 职场话题 · 创业组队 · 外包",
		},
		{
			Key:  "deals",
			Name: "交易",
			Nodes: []string{
				"all4all",
				"exchange",
				"free",
				"dn",
				"tuan",
			},
			NodesTitle: "二手交易 · 物物交换 · 免费赠送 · 域名 · 团购",
		},
		{
			Key:  "city",
			Name: "城市",
			Nodes: []string{
				"beijing",
				"shanghai",
				"shenzhen",
				"guangzhou",
				"hangzhou",
				"chengdu",
				"singapore",
				"nyc",
				"la",
			},
			NodesTitle: "北京 · 上海 · 深圳 · 广州 · 杭州 · 成都 · Singapore · New York · Los Angeles",
		},
		{
			Key:  "qna",
			Name: "问与答",
			Nodes: []string{
				"qna",
			},
			NodesTitle: "问与答",
		},
		{
			Key:  HotNode,
			Name: "最热",
			Nodes: []string{
				HotNode,
			},
			NodesTitle: "hot",
		},
		{
			Key:  LatestNode,
			Name: "最新",
			Nodes: []string{
				LatestNode,
			},
			NodesTitle: "latest",
		},
		{
			Key:   NodesMy,
			Name:  "*我的*",
			Nodes: myNodesKey,
		},
		{
			Key:  "r2",
			Name: "R2",
			Nodes: []string{
				"share",
				"create",
				"qna",
				"jobs",
				"programmer",
				"career",
				"invest",
				"ideas",
				"hardware",
			},
			NodesTitle: "分享发现 · 分享创造 · 问与答 · 酷工作 · 程序员 · 职场话题 · 投资 · 奇思妙想 · 硬件",
		},
		{
			Key:  "vxna",
			Name: "VXNA",
			Nodes: []string{
				"vxna",
				"rss",
				"planet",
				"blogger",
				"webmaster",
			},
			NodesTitle: "VXNA · RSS · Planet · Blogger · 站长",
		},
	}
)

type GroupNode struct {
	Key        string
	Name       string
	Nodes      []string
	NodesTitle string
}

func (g *GroupNode) Title() string {

	if g.Key != NodesMy {
		return g.NodesTitle
	}

	// 是否有过改动
	title := strings.Join(g.Nodes, " · ")
	defaultTitle := strings.Join(myNodesKey, " · ")
	if title == defaultTitle {
		return myNodesTitle
	}

	return title
}

func (g GroupNode) CanChooseApiVersion() bool {

	switch g.Key {
	case NodesMy, LatestNode, HotNode:
		return false
	}

	return true
}
func GetGroupNode(index int) GroupNode {

	groupNode := lo.NthOr(OfficialNodes, index, OfficialNodes[0])
	if groupNode.Key == NodesMy && Config.Get().MyNodes != "" {
		groupNode.Nodes = strings.Split(Config.Get().MyNodes, ",")
	}

	return groupNode
}

func TabNodeIndex(index, add int) int {
	index += add
	if index >= len(OfficialNodes) {
		index = 0
	}

	if index < 0 {
		index = len(OfficialNodes) - 1
	}

	return index
}
