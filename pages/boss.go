package pages

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/model"
)

type bossPage struct {
	packages []string
	index    int
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
}

var (
	ts                  atomic.Int64
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func init() {
	ts.Store(100)
}

func newBossPage() bossPage {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return bossPage{
		packages: getPackages(),
		spinner:  s,
		progress: p,
	}
}

func (m bossPage) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			g.Session.HideFooter.Store(true)
			return nil
		},
		downloadAndInstall(m.packages[m.index]),
		m.spinner.Tick,
	)
}

func (m bossPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.F1):
			return m, func() tea.Msg {
				return g.Config.Save(
					func(config *model.FileConfig) {
						config.BossModeBlank = !config.BossModeBlank
					},
				)
			}
		}
	case installedPkgMsg:
		pkg := m.packages[m.index]
		if m.index >= len(m.packages)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s %s", checkMark, pkg), // print the last success message
				tea.Quit,                            // exit the program
			)
		}
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.packages)))
		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, pkg),     // print success message above our program
			downloadAndInstall(m.packages[m.index]), // download the next package
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m bossPage) View() string {
	n := len(m.packages)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done! Installed %d packages.\n", n))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	pkgName := currentPkgNameStyle.Render(m.packages[m.index])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Installing " + pkgName)

	return lipgloss.NewStyle().Padding(1).Render(spin + info + "\n" + prog + pkgCount)
}

type installedPkgMsg string

func downloadAndInstall(pkg string) tea.Cmd {
	// This is where you'd do i/o stuff to download and install packages. In
	// our case we're just pausing for a moment to simulate the process.
	val := ts.Load()
	ts.Add(val + 100)
	d := time.Millisecond * time.Duration(val)
	return tea.Tick(
		d, func(t time.Time) tea.Msg {
			return installedPkgMsg(pkg)
		},
	)
}

var packages = []string{
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

func getPackages() []string {
	pkgs := packages
	copy(pkgs, packages)

	rand.Shuffle(
		len(pkgs), func(i, j int) {
			pkgs[i], pkgs[j] = pkgs[j], pkgs[i]
		},
	)

	for k := range pkgs {
		pkgs[k] += fmt.Sprintf("-%d.%d.%d", rand.IntN(10), rand.IntN(10), rand.IntN(10)) //nolint:gosec
	}
	return pkgs
}
