package pages

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

const (
	installMaxJobs = 20
)

var (
	count         int
	runningJobs   []string
	packages      = getPackages()
	progressModel = progress.New(
		progress.WithGradient("#636e72", "#2980b9"),
	)
)

type bossPage struct {
	spinner spinner.Model
}

func newBossPage() bossPage {
	return bossPage{
		spinner: spinner.New(spinner.WithSpinner(spinner.Line)),
	}
}

func (m bossPage) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			g.Session.HideFooter.Store(true)
			return nil
		},
		m.spinner.Tick,
		commands.Post(messages.BossInitMsg{}),
	)
}

func (m bossPage) Close() error {
	g.Session.HideFooter.Store(false)
	return nil
}

func (m bossPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.KeyQ):
			return m, commands.RedirectPop()
		case key.Matches(msg, consts.AppKeyMap.F1):
			return m, func() tea.Msg {
				return g.Config.Save(
					func(config *model.FileConfig) {
						config.BossModeBlank = !config.BossModeBlank
					},
				)
			}
		}
	case messages.BossInitMsg:
		var (
			w, _ = g.Window.GetSize()
		)
		progressModel.Width = max(w/2, 20)
		return m, commands.Post(messages.BossStartInstallPkgMsg{})
	case messages.BossStartInstallPkgMsg:
		// 随机安装一个
		for name, _ := range packages {
			return m.downloadAndInstall(name)
		}
		return m, nil
	case messages.BossEndInstallPkgMsg:
		// 任务完成, 标记一下
		p := progressModel.Percent()
		if p > 0.99 {
			return m, tea.Batch(
				progressModel.SetPercent(0),
				commands.Post(messages.BossStartInstallPkgMsg{}),
			)
		}

		// 把 remain 里的删除一下
		return m, tea.Batch(
			progressModel.SetPercent(p+float64(1)/float64(len(packages))),
			commands.Post(messages.BossStartInstallPkgMsg{}),
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := progressModel.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			progressModel = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m bossPage) View() string {

	var (
		content strings.Builder
		// 只保留 5 个显示下载中
		loadingIndex = len(runningJobs) - 5
		checked      = styles.Active.Render("✓")
	)

	for i, name := range runningJobs {
		prefix := checked
		if i > loadingIndex {
			prefix = m.spinner.View()
		}
		content.WriteString(fmt.Sprintf("%s %s", prefix, name))
		content.WriteString("\n")
	}
	content.WriteString(progressModel.View())
	content.WriteString(" ")
	content.WriteString("\n")
	// 显示一个进度条再最上面
	return lipgloss.NewStyle().Padding(1).Render(content.String())
}

func (m bossPage) downloadAndInstall(pkg string) (tea.Model, tea.Cmd) {
	count++
	count = count % 100
	runningJobs = append(runningJobs, pkg)
	if len(runningJobs) > installMaxJobs {
		runningJobs = runningJobs[1:]
	}

	// 基础延迟时间（毫秒）
	baseDelay := 1000.0
	// 根据count值计算延迟倍数
	slowFactor := math.Pow(1.0515, float64(count))
	// 计算最终延迟时间（加入随机因素）
	delay := time.Duration(baseDelay*slowFactor) * time.Millisecond
	return m, tea.Tick(
		delay, func(t time.Time) tea.Msg {
			return messages.BossEndInstallPkgMsg(pkg)
		},
	)
}

func getPackages() map[string]bool {
	items := []string{
		"gin-gonic/gin",
		"labstack/echo",
		"goccy/go-json",
		"go-redis/redis",
		"prometheus/client_golang",
		"grpc/grpc-go",
		"stretchr/testify",
		"gorilla/mux",
		"spf13/cobra",
		"hashicorp/terraform",
		"golang/protobuf",
		"aws/aws-sdk-go",
		"docker/docker",
		"kubernetes/client-go",
		"etcd-io/etcd",
		"influxdata/influxdb",
		"elastic/go-elasticsearch",
		"pingcap/tidb",
		"cockroachdb/cockroach",
		"minio/minio",
		"nats-io/nats-server",
		"mosn.io/mosn",
		"dapr/dapr",
		"istio/istio",
		"helm/helm",
		"argoproj/argo-cd",
		"tektoncd/pipeline",
		"fluxcd/flux2",
		"prometheus/prometheus",
		"grafana/grafana",
		"jaegertracing/jaeger",
		"open-telemetry/opentelemetry-go",
		"uber-go/zap",
		"sirupsen/logrus",
		"go-kit/kit",
		"micro/micro",
		"gorilla/websocket",
		"go-playground/validator",
		"jmoiron/sqlx",
		"go-sql-driver/mysql",
		"lib/pq",
		"mattn/go-sqlite3",
		"mongodb/mongo-go-driver",
		"redis/go-redis/v9",
		"aws/aws-sdk-go-v2",
		"googleapis/google-api-go-client",
		"Azure/azure-sdk-for-go",
		"hashicorp/vault",
		"consul/consul",
		"nomadproject/nomad",
		"pulumi/pulumi",
		"terraform-providers/terraform-provider-aws",
		"go-git/go-git",
		"src-d/go-git",
		"gogs/gogs",
		"gitea/gitea",
		"drone/drone",
		"jenkins-x/jx",
		"tektoncd/cli",
		"knative/serving",
		"buildpacks/pack",
		"containerd/containerd",
		"cri-o/cri-o",
		"opencontainers/runc",
		"helm/chartmuseum",
		"fluxcd/helm-controller",
		"prometheus/node_exporter",
		"prometheus/alertmanager",
		"thanos-io/thanos",
		"cortexproject/cortex",
		"grafana/loki",
		"influxdata/telegraf",
		"trivy/trivy",
		"aquasecurity/trivy",
		"anchore/syft",
		"wagoodman/dive",
		"hadolint/hadolint",
		"golangci/golangci-lint",
		"staticcheck/staticcheck",
		"go-critic/go-critic",
		"uber-go/goleak",
		"dvyukov/go-fuzz",
		"stretchr/objx",
		"mitchellh/mapstructure",
		"spf13/viper",
		"urfave/cli",
		"cobra-cli/cobra",
		"go-yaml/yaml",
		"pelletier/go-toml",
		"json-iterator/go",
		"modern-go/concurrent",
		"modern-go/reflect2",
		"klauspost/compress",
		"google/uuid",
		"hashicorp/golang-lru",
		"alecthomas/kingpin/v2",
		"go-openapi/spec",
		"swaggo/swag",
		"go-swagger/go-swagger",
		"grpc-ecosystem/grpc-gateway",
		"bufbuild/buf",
		"envoyproxy/protoc-gen-validate",
	}
	mutable.Shuffle(items)
	return lo.SliceToMap(
		items, func(item string) (string, bool) {
			return fmt.Sprintf(
				"%s-%d.%d.%d",
				item,
				1+rand.Intn(2),
				1+rand.Intn(30),
				1+rand.Intn(30),
			), true
		},
	)
}
