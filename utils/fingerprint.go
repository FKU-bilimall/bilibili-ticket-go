package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 指纹数据结构
type FingerprintData struct {
	UserAgent           string   `json:"userAgent"`
	ScreenResolution    [2]int   `json:"screenResolution"`
	ColorDepth          int      `json:"colorDepth"`
	TimezoneOffset      int      `json:"timezoneOffset"`
	HardwareConcurrency int      `json:"hardwareConcurrency"`
	DeviceMemory        int      `json:"deviceMemory"`
	TouchSupport        [3]bool  `json:"touchSupport"`
	WebGLVendor         string   `json:"webglVendor"`
	WebGLRenderer       string   `json:"webglRenderer"`
	CanvasFingerprint   string   `json:"canvasFingerprint"`
	AudioFingerprint    string   `json:"audioFingerprint"`
	Fonts               []string `json:"fonts"`
	Plugins             []string `json:"plugins"`
}

// GenerateRandomFingerprint 生成随机指纹数据
func GenerateRandomFingerprint() FingerprintData {
	return FingerprintData{
		UserAgent:           randomUserAgent(),
		ScreenResolution:    randomScreenResolution(),
		ColorDepth:          randomColorDepth(),
		TimezoneOffset:      randomTimezoneOffset(),
		HardwareConcurrency: randomHardwareConcurrency(),
		DeviceMemory:        randomDeviceMemory(),
		TouchSupport:        randomTouchSupport(),
		WebGLVendor:         randomWebGLVendor(),
		WebGLRenderer:       randomWebGLRenderer(),
		CanvasFingerprint:   randomCanvasFingerprint(),
		AudioFingerprint:    randomAudioFingerprint(),
		Fonts:               randomFonts(),
		Plugins:             randomPlugins(),
	}
}

// CalculateFingerprintID 生成指纹ID(使用SHA256哈希)
func CalculateFingerprintID(fp FingerprintData) string {
	data := fmt.Sprintf("%v", fp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// 随机UserAgent生成
func randomUserAgent() string {
	browsers := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:%d.0) Gecko/20100101 Firefox/%d.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%d.%d Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.%d.%d Safari/537.36 Edg/%d.0.%d.%d",
	}

	template := browsers[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(browsers))]
	switch {
	case strings.Contains(template, "Chrome"):
		return fmt.Sprintf(template, 90+rand.Intn(20), rand.Intn(1000), rand.Intn(100), 90+rand.Intn(20), rand.Intn(1000))
	case strings.Contains(template, "Firefox"):
		ver := 90 + rand.Intn(20)
		return fmt.Sprintf(template, ver, ver)
	case strings.Contains(template, "Safari"):
		return fmt.Sprintf(template, 14+rand.Intn(5), rand.Intn(10))
	case strings.Contains(template, "Edg"):
		return fmt.Sprintf(template, 90+rand.Intn(20), rand.Intn(1000), rand.Intn(100), 90+rand.Intn(20), rand.Intn(1000), rand.Intn(100))
	default:
		return browsers[0]
	}
}

// 其他随机生成函数...
func randomScreenResolution() [2]int {
	resolutions := [][2]int{
		{1920, 1080}, {1366, 768}, {1440, 900},
		{1536, 864}, {1600, 900}, {1280, 720},
		{2560, 1440}, {3840, 2160}, {1024, 768},
	}
	return resolutions[rand.Intn(len(resolutions))]
}

func randomColorDepth() int {
	return 24 // 大多数现代设备使用24位色深
}

func randomTimezoneOffset() int {
	return -720 + rand.Intn(1440) // -12到+12小时范围
}

func randomHardwareConcurrency() int {
	return 2 << rand.Intn(4) // 2,4,8,16核
}

func randomDeviceMemory() int {
	return 2 << rand.Intn(4) // 2,4,8,16GB
}

func randomTouchSupport() [3]bool {
	return [3]bool{
		rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2) == 1, // maxTouchPoints
		rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2) == 1, // touchEvent
		rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2) == 1, // ontouchstart
	}
}

func randomWebGLVendor() string {
	vendors := []string{
		"Google Inc.", "Intel Inc.", "NVIDIA Corporation",
		"AMD", "Apple Inc.", "Microsoft",
	}
	return vendors[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(vendors))]
}

func randomWebGLRenderer() string {
	renderers := []string{
		"ANGLE (Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA GeForce GTX 1060 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (AMD Radeon RX 580 Direct3D11 vs_5_0 ps_5_0)",
		"Apple GPU", "Mali-G72", "Adreno (TM) 630",
	}
	return renderers[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(renderers))]
}

func randomCanvasFingerprint() string {
	// 模拟Canvas指纹的base64数据前缀
	return "data:image/png;base64," + randomHexString(32)
}

func randomAudioFingerprint() string {
	// 模拟音频指纹
	return fmt.Sprintf("%d.%d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000), rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000))
}

func randomFonts() []string {
	fonts := []string{
		"Arial", "Times New Roman", "Courier New",
		"Georgia", "Verdana", "Helvetica",
		"Tahoma", "Calibri", "Cambria",
	}

	// 随机选择3-8种字体
	count := 3 + rand.New(rand.NewSource(time.Now().UnixNano())).Intn(6)
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fonts[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(fonts))]
	}
	return result
}

func randomPlugins() []string {
	plugins := []string{
		"Chrome PDF Viewer",
		"Native Client",
		"Widevine Content Decryption Module",
		"Microsoft Edge PDF Viewer",
		"WebKit built-in PDF",
	}

	// 随机选择1-4个插件
	count := 1 + rand.New(rand.NewSource(time.Now().UnixNano())).Intn(4)
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = plugins[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(plugins))]
	}
	return result
}

func randomHexString(length int) string {
	b := make([]byte, length/2)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(b)
	return hex.EncodeToString(b)
}
