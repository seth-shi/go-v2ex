package api

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	mockResponseCount atomic.Int64
)

func mockApiResp(req *http.Request, resp *http.Response) {
	delay := time.Duration(rand.Intn(1)+1) * time.Second
	time.Sleep(delay)
	var (
		code            = http.StatusNotFound
		key             = safeUrlToFileKey(req)
		getBody, exists = mockResponse[key]
		body            string
	)
	slog.Info(
		"mock_response",
		slog.String("key", key),
		slog.Bool("exists", exists),
	)
	if exists {
		code = http.StatusOK
		body = getBody()
	}

	respMockJson(resp, code, body)
}

func respMockJson(resp *http.Response, code int, body string) *http.Response {
	mockResponseCount.Add(1)
	maxCount := 100
	remain := maxCount - int(mockResponseCount.Load())
	if remain < 0 {
		remain = 0
	}

	resp.StatusCode = code
	resp.Body = io.NopCloser(strings.NewReader(body))
	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set(headerLimit, strconv.Itoa(maxCount))
	resp.Header.Set(headerRemain, strconv.Itoa(remain))
	resp.ContentLength = int64(len(body))
	return resp
}

func safeUrlToFileKey(r *http.Request) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"?", "_",
		"=", "_",
	)
	return fmt.Sprintf(
		"%s_%s.json",
		replacer.Replace(strings.TrimLeft(r.URL.Path, "/")),
		replacer.Replace(r.URL.Query().Encode()),
	)
}

type getMockResponse func() string

var (
	// 数据来自: https://github.com/Eminlin/V2Ei
	mockResponse = map[string]getMockResponse{
		"api_topics_hot.json_.json":             mockHotTopics,
		"api_v2_nodes_qna_topics_p_1.json":      mockQnaTopics,
		"api_v2_token_.json":                    mockToken,
		"api_v2_member_.json":                   mockMember,
		"api_v2_topics_560297_.json":            mockDetail,
		"api_v2_topics_560297_replies_p_1.json": mockReply(1),
		"api_v2_topics_560297_replies_p_2.json": mockReply(2),
		"api_v2_topics_560297_replies_p_3.json": mockReply(3),
	}
)

func mockReply(page int) getMockResponse {
	return func() string {
		switch page {
		case 1:
			return `{"result":[{"id":7264186,"content":"模拟器不错，命令行串流太天才了（不知道 FPS 有没有 5 …","created":1556675689,"member":{"id":7176,"username":"oott123"}},{"id":7264190,"content":"@oott123 命令行我锁了 10fps，网络理想情况下大概 5-10 之间吧（肉眼观测）","created":1556675745,"member":{"id":250736,"username":"AaronLiu00"}},{"id":7264192,"content":"呵呵呵就这破玩意，不是我吹，给我五百年我也整不明白","created":1556675787,"member":{"id":340004,"username":"BreezeInWind"}},{"id":7264202,"content":"好顶赞！ 已 star","created":1556676013,"member":{"id":101604,"username":"0312birdzhang"}},{"id":7264213,"content":"厉害厉害 支持一下","created":1556676254,"member":{"id":238172,"username":"pakro888"}},{"id":7264272,"content":"才大二就做出这种东西好厉害啊，我大二的时候他妈都在干什么。已星","created":1556677408,"member":{"id":181466,"username":"shihira"}},{"id":7264299,"content":"支持！!","created":1556678010,"member":{"id":336784,"username":"Mayuri"}},{"id":7264311,"content":"@hedamao9999 \r\n@0312birdzhang \r\n@pakro888 \r\n@shihira \r\n@Mayuri \r\n\r\n感谢支持~","created":1556678160,"member":{"id":250736,"username":"AaronLiu00"}},{"id":7264343,"content":"嘻嘻，喜欢你的树洞外链，很方便\n一看头像就知道是你～","created":1556678789,"member":{"id":307887,"username":"vanishcode"}},{"id":7264446,"content":"大佬","created":1556680777,"member":{"id":51170,"username":"Tink"}},{"id":7264464,"content":"真是个人才，已 star，研究一下能不能回味一下水浒神兽","created":1556681088,"member":{"id":33440,"username":"zwpaper"}},{"id":7264503,"content":"我大二在干嘛...\r\n在玩骨头🦴","created":1556682076,"member":{"id":288843,"username":"acupnocup"}},{"id":7264514,"content":"膜拜大佬","created":1556682307,"member":{"id":12297,"username":"jon"}},{"id":7264527,"content":"大佬大佬 居然才大二","created":1556682502,"member":{"id":239634,"username":"YuuuZeee"}},{"id":7264537,"content":"大佬","created":1556682593,"member":{"id":114136,"username":"abmin521"}},{"id":7264579,"content":"很厉害 支持支持","created":1556683525,"member":{"id":235324,"username":"Doodlister"}},{"id":7264639,"content":"厉害","created":1556684798,"member":{"id":153086,"username":"zhihaofans"}},{"id":7264739,"content":"哈哈哈标签页很皮啊   是怎么实现的呢","created":1556686923,"member":{"id":91129,"username":"isnowify"}},{"id":7264933,"content":"@isnowify 标签页指的是？","created":1556691888,"member":{"id":250736,"username":"AaronLiu00"}},{"id":7264935,"content":"厉害","created":1556691948,"member":{"id":115781,"username":"kidtest"}}],"pagination":{"total":42,"pages":3}}`
		case 2:
			return `{"result":[{"id":7264971,"content":"@AaronLiu00 切换到其他选项卡后 你的网站标题栏会故意乱码","created":1556692756,"member":{"id":91129,"username":"isnowify"}},{"id":7264976,"content":"@isnowify 哦哦这个 可以参考 https://diygod.me/2153/","created":1556692857,"member":{"id":250736,"username":"AaronLiu00"}},{"id":7265016,"content":"太厉害了，赶紧 star","created":1556694198,"member":{"id":86866,"username":"tony601818"}},{"id":7265038,"content":"厉害，star","created":1556694742,"member":{"id":370010,"username":"zuokanyunqishi"}},{"id":7265435,"content":"这也太强了","created":1556705233,"member":{"id":294163,"username":"Mantext1989"}},{"id":7265440,"content":"nb","created":1556705352,"member":{"id":197959,"username":"Mystic"}},{"id":7265461,"content":"厉害，star 了","created":1556706084,"member":{"id":194317,"username":"Archeb"}},{"id":7265552,"content":"Star","created":1556708314,"member":{"id":257835,"username":"FDKevin"}},{"id":7265569,"content":"居然是命令行显示，lz 厉害了","created":1556709011,"member":{"id":314537,"username":"daweii"}},{"id":7265592,"content":"“但是我偶然间看到一篇关于 Gameboy 模拟器的 Tutorial ”\r\n\r\n能发一下地址吗","created":1556709884,"member":{"id":314537,"username":"daweii"}},{"id":7265598,"content":"@daweii Reference 下面的几个都挺不错的。我这里指的是这个： http://www.codeslinger.co.uk/pages/projects/gameboy/beginning.html","created":1556710162,"member":{"id":250736,"username":"AaronLiu00"}},{"id":7265854,"content":"好强。。佩服","created":1556717316,"member":{"id":277632,"username":"zhanwh9"}},{"id":7266343,"content":"C...Cloudreve","created":1556739107,"member":{"id":396659,"username":"Ayersneo"}},{"id":7266474,"content":"恭喜上榜 Github Trending","created":1556755986,"member":{"id":257752,"username":"kyokuheishin"}},{"id":7266485,"content":"@kyokuheishin 一觉醒来，看到这个惊呆了😂","created":1556756454,"member":{"id":250736,"username":"AaronLiu00"}},{"id":7266814,"content":"太强了","created":1556765213,"member":{"id":343648,"username":"YiferHuang"}},{"id":7267825,"content":"哈哈哈，这个有意思！ star 了！","created":1556789101,"member":{"id":309696,"username":"Dawnki"}},{"id":7269650,"content":"什么是串流？(真诚问","created":1556854252,"member":{"id":402378,"username":"good1uck"}},{"id":7271331,"content":"厉害，已推荐到微博 https://weibo.com/5722964389/Hsrt1vVMt","created":1556897327,"member":{"id":374541,"username":"GitHubDaily"}},{"id":7275093,"content":"前段时间还在看 JS 的 nes 模拟器。\r\n没想到这个更厉害。","created":1557020152,"member":{"id":398794,"username":"zhensjoke"}}],"pagination":{"total":42,"pages":3}}`
		case 3:
			return `{"result":[{"id":7294630,"content":"NB","created":1557301461,"member":{"id":182556,"username":"sailei"}},{"id":7460365,"content":"人才啊 ...","created":1560152323,"member":{"id":52447,"username":"1ychee"}}],"pagination":{"total":42,"pages":3}}`
		}
		return `{}`
	}
}

func mockDetail() string {
	return `
{
    "result": {
        "id": 560297,
        "title": "尝试写了一个 Gameboy 模拟器，支持在命令行下“云游戏串流”游玩",
        "content": "# 效果\r\n\r\n传统的 Gameboy 游戏模拟：\r\n\r\nhttps://s2.ax1x.com/2019/05/01/EJV1sO.png\r\n\r\n当然，正如标题描述，只需要一条命令，无需额外安装软件，你就能在命令行下游玩 Gameboy 游戏了：\r\n\r\n\r\ntelnet gameboy.live 1989\r\n\r\n\r\nhttps://s2.ax1x.com/2019/05/01/EJVJdH.jpg\r\n\r\n要注意的是，云游戏只能在支持 ANSI 标准和 UTF-8 编码的终端下游玩。Windows 下可以在 WSL 里玩。如果提示命令未找到，安装 telnet 就行了。\r\n\r\n# 源代码\r\n\r\nGitHub： https://github.com/HFO4/gameboy.live\r\n\r\n(刚好赶上成为平成最后的 Gameboy 模拟器)\r\n\r\n# 为什么要写这个，以及一些体会...\r\n\r\n这个项目呢，并不是为了模拟器本身，毕竟更完善更稳定的模拟器有不少。完成这个项目更偏向是自我学习吧，楼主目前大二，上学期刚学了汇编和计组，老师也劝我们大学期前写一点成型的项目出来，再加上我是个任天堂粉丝，虽然没有经历过那个时代，但又对老式家用机和掌机有着额外的兴趣，特别是 Gameboy。给 Gameboy 写模拟器一直算是我的一个梦想吧，之前也稍微研究过 Gameboy 通信接口（有关相机和打印机外设的，有兴趣的可以去看下之前写的文章：[用树莓派模拟 Game Boy 打印机及相机外设]( https://aoaoao.me/2018/12/17/game-boy-printer/)），对 Gameboy 硬件有了基本的了解，那个时候突然发现用刚学的计组好像...可以对模拟器原理理解个大概了，然后就跳入了这个深坑。\r\n\r\n开始写代码之前我构思了很久，虽然大概理解了基本结构，但是具体的实现还是无从下手。但是我偶然间看到一篇关于 Gameboy 模拟器的 Tutorial，看完后感到醍醐灌顶，思路上就很清晰了。\r\n\r\n真正写代码的过程，真的一言难尽。大体上就是写半小时代码，Debug 一整天。模拟器这玩意儿 Debug 起来挺麻烦的，我采用的办法是和其他模拟器对比，单步执行每条指令，在对比寄存器和各种状态，缩小锁定出现偏差的位置。有好几天我在梦里都在用人脑模拟 CPU，基本上除了上课吃饭睡觉，别的时间都在搞这个了 QAQ  最难的部分不是 CPU，也不是图形，而是声音的模拟。因为没有相关知识储备，看着文档里的 envelope sweep 这些词不知所措。弄了好久最后终于算是能听的级别了，但是跟真机相比还是有区别。\r\n\r\n总的来说写这个收获真的很大，原本以为用不到的汇编和计组课程知识在这里也派上了用场。第一次看到游戏画面展示出来的那一刻，真的很爽。",
        "url": "https://www.v2ex.com/t/560297",
        "replies": 42,
        "last_reply_by": "1ychee",
        "created": 1556675267,
        "last_modified": 1556675267,
        "last_touched": 1557363927,
        "member": {
            "id": 250736,
            "username": "AaronLiu00"
        },
        "node": {
            "id": 17,
            "name": "create",
            "title": "分享创造"
        },
        "supplements": []
    }
}
`
}

func mockMember() string {
	return `
{
  "success": true, 
  "message": "Current token details", 
  "result": {
    "id": 1
  }
}`
}

func mockToken() string {

	now := time.Now().Unix()
	exp := time.Now().Unix() + 600
	return fmt.Sprintf(
		`
{
  "success": true, 
  "message": "Current token details", 
  "result": {
    "token": "00000000-0000-0000-0000-000000000000", 
    "scope": "everything", 
    "expiration": %d, 
    "good_for_days": 100, 
    "total_used": 1, 
    "last_used": %d, 
    "created": %d
  }
}
`,
		exp,
		now,
		now,
	)
}

func mockQnaTopics() string {
	return `
{
    "pagination": {
        "total": 20,
        "pages": 2
    },
    "result": [
        {
            "id": 560297,
            "title": "尝试写了一个 Gameboy 模拟器，支持在命令行下“云游戏串流”游玩",
            "replies": 42,
            "last_reply_by": "AaronLiu00",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1530502042,
            "last_touched": 1640635969
        },
        {
            "id": 467407,
            "title": "分享个自用的小工具~ 给你的 iPhone 发自定义推送",
            "replies": 218,
            "last_reply_by": "finab",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1530502041,
            "last_touched": 1640635968
        },
        {
            "id": 555768,
            "title": "Tea + Cloud，那个为开发者而生的笔记应用，它上天（云）了！",
            "replies": 33,
            "last_reply_by": "hk3475",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1555406874,
            "last_touched": 1557362740
        },
        {
            "id": 553321,
            "title": "利用公交线路可视化城市结构",
            "replies": 60,
            "last_reply_by": "96486d9b",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1554785871,
            "last_touched": 1561703564
        },
        {
            "id": 532913,
            "title": "老爹的铁铺上线，给老爸做个广告，云上铁铺 :)",
            "replies": 166,
            "last_reply_by": "bokchoys",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1549208966,
            "last_touched": 1555816164
        },
        {
            "id": 574208,
            "title": "歪国程序员脑洞真的不是一般的大，这次他们要在 URL 上打游戏！😂",
            "replies": 10,
            "last_reply_by": "keelii",
            "node": {
                "id": 519,
                "name": "ideas",
                "title": "奇思妙想"
            },
            "created": 1560581389,
            "last_touched": 1560629212
        },
        {
            "id": 561958,
            "title": "我想开发一门新的编程语言，不过个人能力有限（编程技术很菜ヾ(ｏ･ω･)ﾉ，不过并不影响我对编程语言的理解），希望有人帮助我开发编译器或解释器，完整的想法我已经有了，就等实现了。",
            "replies": 319,
            "last_reply_by": "Qiaogui",
            "node": {
                "id": 300,
                "name": "programmer",
                "title": "程序员"
            },
            "created": 1557232215,
            "last_touched": 1561906875
        },
        {
            "id": 549223,
            "title": "NVIDIA 基于自家 Jetson Nano 开源机器人 Jetbot DIY 资料汇总",
            "replies": 7,
            "last_reply_by": "unbug",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1553678657,
            "last_touched": 1553824068
        },
        {
            "id": 541987,
            "title": "各大网站登陆方式， 包括爬虫，麻麻再也不用担心我学习爬虫啦，哈哈",
            "replies": 118,
            "last_reply_by": "CriseLYJ",
            "node": {
                "id": 90,
                "name": "python",
                "title": "Python"
            },
            "created": 1551924078,
            "last_touched": 1552690404
        },
        {
            "id": 548519,
            "title": "工作三到五年后接触机器学习的入门建议",
            "replies": 43,
            "last_reply_by": "theworldsong",
            "node": {
                "id": 678,
                "name": "ml",
                "title": "机器学习"
            },
            "created": 1553529587,
            "last_touched": 1553610832
        },
        {
            "id": 675067,
            "title": "程序员就一定要去 IT 公司工作吗？",
            "replies": 141,
            "last_reply_by": "clockOS",
            "node": {
                "id": 300,
                "name": "programmer",
                "title": "程序员"
            },
            "created": 1590365266,
            "last_touched": 1590437790
        },
        {
            "id": 552627,
            "title": "开源肖像-向伟大的开源领袖们致敬",
            "replies": 2,
            "last_reply_by": "bigezhang",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1554603678,
            "last_touched": 1554509105
        },
        {
            "id": 611963,
            "title": "数据结构在实际项目中的使用 - 链表",
            "replies": 11,
            "last_reply_by": "gansteed",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1571794342,
            "last_touched": 1571811447
        },
        {
            "id": 211400,
            "title": "收集 V2EX 上的撕逼大战",
            "replies": 95,
            "last_reply_by": "greatghoul",
            "node": {
                "id": 148,
                "name": "pointless",
                "title": "无要点"
            },
            "created": 1438907585,
            "last_touched": 1448297364
        },
        {
            "id": 550812,
            "title": "你们的启蒙编程语言是？",
            "replies": 456,
            "last_reply_by": "szzhiyang",
            "node": {
                "id": 300,
                "name": "programmer",
                "title": "程序员"
            },
            "created": 1554098708,
            "last_touched": 1554510431
        },
        {
            "id": 550681,
            "title": "前端萌新正在做的中国风 React 组件库...",
            "replies": 97,
            "last_reply_by": "AddOneG",
            "node": {
                "id": 17,
                "name": "create",
                "title": "分享创造"
            },
            "created": 1554080893,
            "last_touched": 1554381922
        },
        {
            "id": 567774,
            "title": "如果有云电脑这种东西 你们会使用吗",
            "replies": 115,
            "last_reply_by": "titadida",
            "node": {
                "id": 519,
                "name": "ideas",
                "title": "奇思妙想"
            },
            "created": 1558853689,
            "last_touched": 1559899895
        },
        {
            "id": 695254,
            "title": "大家有没有坚持了很久的观点或者想法，突然发现是自己错了",
            "replies": 151,
            "last_reply_by": "minglanyu",
            "node": {
                "id": 12,
                "name": "qna",
                "title": "问与答"
            },
            "created": 1596438101,
            "last_touched": 1596533481
        },
        {
            "id": 585301,
            "title": "假如有一天脑机接口真的实现了，意识可以被存储甚至复制，那么人类是否可以永生？",
            "replies": 133,
            "last_reply_by": "maxxfire",
            "node": {
                "id": 519,
                "name": "idea",
                "title": "奇思妙想"
            },
            "created": 1563847452,
            "last_touched": 1563943915
        },
        {
            "id": 574173,
            "title": "让你在家，在办公室，在任何地方听到森林，溪流的声音",
            "replies": 23,
            "last_reply_by": "bokchoys",
            "node": {
                "id": 519,
                "name": "ideas",
                "title": "奇思妙想"
            },
            "created": 1560571804,
            "last_touched": 1560606793
        }
    ]
}
`
}

func mockHotTopics() string {
	return `
[
  {
    "id": 490730,
    "title": "JS 不写分号会出 BUG 的。。。",
    "replies": 97,
    "member": {
      "id": 209078,
      "username": "fundebug",
	  "pro": 1
    },
    "node": {
      "id": 146,
      "name": "js",
      "title": "JavaScript"
    },
    "created": 1537326783,
    "last_touched": 1537718856
  },
  {
    "id": 331037,
    "title": "来来来，学 Laraval 的，学 TP 的给你出一道简单的题",
    "replies": 168,
    "member": {
      "id": 88451,
      "username": "abc123ccc"
    },
    "node": {
      "id": 314,
      "name": "flamewar",
      "title": "水深火热"
    },
    "created": 1482999162,
    "last_touched": 1483971983
  },
  {
    "id": 337917,
    "title": "SpaceVim 和 Space-Vim 哪个才是真的？",
    "replies": 116,
    "member": {
      "id": 206200,
      "username": "coderzys"
    },
    "node": {
      "id": 249,
      "name": "vim",
      "title": "Vim"
    },
    "created": 1486120848,
    "last_touched": 1506570270
  },
  {
    "id": 222843,
    "title": "100offer 刷票脚本",
    "replies": 26,
    "member": {
      "id": 28986,
      "username": "vitovan"
    },
    "node": {
      "id": 300,
      "name": "programmer",
      "title": "程序员"
    },
    "created": 1442912607,
    "last_touched": 1442921078
  },
  {
    "id": 223217,
    "title": "亲眼见证了 100 offer 的那个活动的一个项目，半小时内从 0 到 9000 票",
    "replies": 76,
    "member": {
      "id": 53761,
      "username": "lincanbin"
    },
    "node": {
      "id": 300,
      "name": "programmer",
      "title": "程序员"
    },
    "created": 1443022531,
    "last_touched": 1443157764
  },
  {
    "id": 538109,
    "title": "其实中国如果想在未来一代人身上大幅降低癌症发病率也很简单，做到以下几条就可以：",
    "replies": 174,
    "member": {
      "id": 19665,
      "username": "ccming"
    },
    "node": {
      "id": 700,
      "name": "fit",
      "title": "健康"
    },
    "created": 1550972727,
    "last_touched": 1551730403
  },
  {
    "id": 360207,
    "title": "看完这几个人的经历，你还会提倡“知识付费”吗？",
    "replies": 104,
    "member": {
      "id": 127148,
      "username": "chlo0823"
    },
    "node": {
      "id": 12,
      "name": "qna",
      "title": "问与答"
    },
    "created": 1494329166,
    "last_touched": 1494410669
  },
  {
    "id": 281794,
    "title": "母亲来上海照顾小孩，但她觉得孤独，基本没有什么社交，你们是怎么处理呢？",
    "replies": 164,
    "member": {
      "id": 84957,
      "username": "metrue"
    },
    "node": {
      "id": 18,
      "name": "shanghai",
      "title": "上海"
    },
    "created": 1464360601,
    "last_touched": 1464872188
  },
  {
    "id": 359798,
    "title": "我只想说跑去二楼进电梯的跟插队没区别 ! 鄙视.",
    "replies": 298,
    "member": {
      "id": 200500,
      "username": "371657110"
    },
    "node": {
      "id": 380,
      "name": "flood",
      "title": "水"
    },
    "created": 1494208171,
    "last_touched": 1494413419
  },
  {
    "id": 327127,
    "title": "天猫真的是匿名评论吗？",
    "replies": 85,
    "member": {
      "id": 129081,
      "username": "cocacold"
    },
    "node": {
      "id": 12,
      "name": "qna",
      "title": "问与答"
    },
    "created": 1481539078,
    "last_touched": 1482584239
  },
  {
    "id": 195162,
    "title": "煎蛋被扒站这事大家怎么看？",
    "replies": 91,
    "member": {
      "id": 464,
      "username": "underone"
    },
    "node": {
      "id": 12,
      "name": "qna",
      "title": "问与答"
    },
    "created": 1433085235,
    "last_touched": 1436216852
  },
  {
    "id": 256781,
    "title": "程序员做微商并不丢人。",
    "replies": 112,
    "member": {
      "id": 93055,
      "username": "AmberBlack"
    },
    "node": {
      "id": 314,
      "name": "flamewar",
      "title": "水深火热"
    },
    "created": 1455592945,
    "last_touched": 1455724891
  },
  {
    "id": 169445,
    "title": "买了锤子后极度后悔啊咋整?",
    "replies": 171,
    "member": {
      "id": 84029,
      "username": "rockybi"
    },
    "node": {
      "id": 314,
      "name": "flamewar",
      "title": "水深火热"
    },
    "created": 1423449198,
    "last_touched": 1423603438
  },
  {
    "id": 443087,
    "title": "关于 2018 年 3 月 31 日遇到的假毕业证书垃圾信息刷屏",
    "replies": 173,
    "member": {
      "id": 1,
      "username": "Livid"
    },
    "node": {
      "id": 2,
      "name": "v2ex",
      "title": "V2EX"
    },
    "created": 1522467224,
    "last_touched": 1529140803
  },
  {
    "id": 243125,
    "title": "HR 果然是很恐怖的一类人",
    "replies": 81,
    "member": {
      "id": 96624,
      "username": "lijianying10"
    },
    "node": {
      "id": 770,
      "name": "career",
      "title": "职场话题"
    },
    "created": 1449934695,
    "last_touched": 1453768455
  },
  {
    "id": 369422,
    "title": "曾经的少年黑客，准备创业了。你们怎么看？",
    "replies": 144,
    "member": {
      "id": 223951,
      "username": "lidongwei"
    },
    "node": {
      "id": 314,
      "name": "flamewar",
      "title": "水深火热"
    },
    "created": 1497801481,
    "last_touched": 1498015123
  },
  {
    "id": 244145,
    "title": "V2EX 把我的邮箱卖了？",
    "replies": 143,
    "member": {
      "id": 58831,
      "username": "riaqn"
    },
    "node": {
      "id": 96,
      "name": "feedback",
      "title": "反馈"
    },
    "created": 1450321435,
    "last_touched": 1452580966
  },
  {
    "id": 77921,
    "title": "极路由窃取用户信息，诸位小心！",
    "replies": 212,
    "member": {
      "id": 3374,
      "username": "chainkhoo"
    },
    "node": {
      "id": 314,
      "name": "flamewar",
      "title": "水深火热"
    },
    "created": 1375580086,
    "last_touched": 1376544868
  },
  {
    "id": 29091,
    "title": "MBP免费送，国行全新，绝不标题党",
    "replies": 282,
    "member": {
      "id": 17542,
      "username": "0xTao"
    },
    "node": {
      "id": 380,
      "name": "flood",
      "title": "水"
    },
    "created": 1331213702,
    "last_touched": 1342256199
  },
  {
    "id": 200500,
    "title": "@Huadb 被差评了，我是来道歉的（前后无关）",
    "replies": 34,
    "member": {
      "id": 21674,
      "username": "manoon"
    },
    "node": {
      "id": 314,
      "name": "flamewar",
      "title": "水深火热"
    },
    "created": 1435030330,
    "last_touched": 1435064688
  }
]
`
}
