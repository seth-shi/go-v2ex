package pkg

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/model"
	"resty.dev/v3"
)

//go:embed data/gamelive.png
var mockImageData []byte

// 正则表达式匹配 http://i.imgur.com/xxxx 或者 http://i.imgur.com/xxxx.png 格式的链接
var (
	imagePattern = regexp.MustCompile(`https?://\S+?\.(?:png|jpe?g)\b`)
	imgClient    *resty.Client
)

func SetUpImageHttpClient(conf *model.FileConfig) {
	imgClient = NewHTTPClient(conf)
	imgClient.
		SetDoNotParseResponse(true).
		AddRequestMiddleware(withImageRequestUserAgent()).
		SetLogger(RestyLogger())

	if conf.IsMockEnv() {
		mock := &MockRoundTripper{
			Mock: func(req *http.Request, resp *http.Response) {
				time.Sleep(time.Second)
				resp.Header.Set("Content-Type", "image/png")
				resp.Body = io.NopCloser(bytes.NewReader(mockImageData))
				resp.ContentLength = int64(len(mockImageData))
			},
		}
		imgClient.SetTransport(mock)
	}
}

func withImageRequestUserAgent() resty.RequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {

		urls, err := url.Parse(req.URL)
		if err != nil {
			return err
		}

		req.Header.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		)
		req.Header.Set("Accept", "image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
		req.Header.Set("Referer", fmt.Sprintf("%s://%s", urls.Scheme, urls.Host))
		return nil
	}
}

func ExtractImgURLsNoUnique(content string) []string {
	return imagePattern.FindAllString(content, -1)
}

func ExtractImgURLs(content string) []string {
	// 合并 imgurPattern 和新的 imagePattern 的匹配结果
	imageRe := imagePattern.FindAllString(content, -1)
	return lo.Uniq(imageRe)
}

type imgRes struct {
	URL  string
	Data image.Image
	err  error
}

func ProcessURLs(urls []string, width int) map[string]string {
	// 处理每个URL
	var (
		wg          sync.WaitGroup
		chSemaphore = make(chan struct{}, 5)
		chImgRes    = make(chan imgRes)
	)

	go func() {
		for _, l := range urls {
			chSemaphore <- struct{}{}
			wg.Add(1)
			go downloadImageRes(l, chSemaphore, chImgRes, &wg)
		}

		wg.Wait()
		close(chImgRes)
	}()

	// 从缓存中获取数据
	var (
		result = make(map[string]string)
	)
	for val := range chImgRes {
		if val.err != nil {
			slog.Error("图片处理失败", slog.String("url", val.URL), slog.Any("err", val.err))
			result[val.URL] = ""
			continue
		}

		slog.Info("图片信息", slog.String("size", val.Data.Bounds().Size().String()), slog.String("url", val.URL))
		str, err := imageToAnsImage(width, val.Data)
		if err != nil {
			slog.Error("图片转字符失败", slog.String("url", val.URL), slog.Any("err", err))
			result[val.URL] = ""
			continue
		}

		slog.Info("图片处理完成", slog.Int("size", len(str)), slog.String("url", val.URL))
		result[val.URL] = str
	}

	// 等待所有任务完成
	return result
}

// 处理单个图片
func downloadImageRes(imgUrl string, semaphore chan struct{}, res chan imgRes, wg *sync.WaitGroup) {

	defer func() {
		<-semaphore
		wg.Done()
	}()

	var (
		data = imgRes{URL: imgUrl}
	)

	// 下载图片
	resp, err := imgClient.
		R().
		Get(imgUrl)
	if err != nil {
		data.err = err
		res <- data
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header().Get("Content-Type")
	if !strings.Contains(contentType, "image") {
		data.err = fmt.Errorf("响应头不是图片:%s", contentType)
		res <- data
		return
	}

	decodeImg, _, err := image.Decode(resp.Body)
	if err != nil {
		data.err = fmt.Errorf("解码图片失败:%+v", err)
		res <- data
		return
	}

	res <- imgRes{
		URL:  imgUrl,
		Data: decodeImg,
	}
}

func imageToAnsImage(width int, img image.Image) (data string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			slog.Error("imageToString panic", slog.Any("err", err))
		}
	}()

	var (
		imgData *ansimage.ANSImage
	)
	// 宽度
	imageWidth := img.Bounds().Max.X - img.Bounds().Min.X
	imageWidth /= 10
	if imageWidth < width {
		width = imageWidth
	}
	// 表情包
	if imageWidth < 10 {
		width = 20
	}

	slog.Info("图片开始渲染", slog.Int("width", width))
	imgData, err = ansimage.NewScaledFromImage(
		img,
		0,
		width,
		color.White,
		ansimage.ScaleModeResize,
		ansimage.NoDithering,
	)
	if err != nil {
		err = fmt.Errorf("ansimage:%+v", err.Error())
		return
	}

	slog.Info("图片渲染完成", slog.Int("width", imgData.Width()))

	return imgData.Render(), nil
}
