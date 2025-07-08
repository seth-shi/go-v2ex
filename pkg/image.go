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
	"github.com/seth-shi/go-v2ex/model"
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
	Data string
}

func ProcessURLs(urls []string, width int) map[string]string {
	// 处理每个URL
	var (
		wg          sync.WaitGroup
		chSemaphore = make(chan struct{}, 5)
		chImgRes    = make(chan imgRes)
	)
	for _, l := range urls {
		wg.Add(1)
		// 限制并发量
		chSemaphore <- struct{}{}
		go processImage(l, width, chSemaphore, chImgRes, &wg)
	}
	go func() {
		wg.Wait()
		close(chImgRes)
	}()

	// 从缓存中获取数据
	var (
		result = make(map[string]string)
	)
	for val := range chImgRes {
		slog.Info("处理图片完成", slog.String("url", val.URL), slog.String("image", val.Data))
		result[val.URL] = val.Data
	}

	// 等待所有任务完成
	return result
}

// 处理单个图片
func processImage(imgUrl string, width int, semaphore chan struct{}, res chan imgRes, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("processImage panic", slog.String("err", fmt.Sprintf("%v", r)))
		}

		<-semaphore
		wg.Done()
	}()

	data, err := imageToString(imgUrl, width)
	if err != nil {
		slog.Error("图片转字符失败", slog.Any("err", lo.Substring(err.Error(), 0, 50)))
		return
	}

	res <- imgRes{
		URL:  imgUrl,
		Data: data,
	}
}

func imageToString(
	imgUrl string,
	width int,
) (data string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("imageToString panic", slog.String("err", fmt.Sprintf("%v", r)))
		}
	}()

	var (
		imgData *ansimage.ANSImage
	)

	res, err := imgClient.
		R().
		Get(imgUrl)

	if err != nil {
		return "", fmt.Errorf("图片下载失败:%s", err.Error())
	}

	defer res.Body.Close()

	contentType := res.Header().Get("Content-Type")
	if !strings.Contains(contentType, "image") {
		return "", fmt.Errorf("响应头不是图片:%s", contentType)
	}

	decodeImg, _, err := image.Decode(res.Body)
	if err != nil {
		return "", fmt.Errorf("解码图片失败:%+v", err)
	}

	// 宽度
	imageWidth := decodeImg.Bounds().Max.X - decodeImg.Bounds().Min.X
	imageWidth /= 10
	slog.Info("图片宽度", slog.Int("width", width), slog.Int("imageWidth", imageWidth))
	if imageWidth < width {
		width = imageWidth
	}
	imgData, err = ansimage.NewScaledFromImage(
		decodeImg,
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

	data = imgData.Render()
	return
}
