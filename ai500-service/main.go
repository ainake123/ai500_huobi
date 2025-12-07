package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ================= 配置 =================

const (
	ServerPort  = ":2234"
	HuobiAPI    = "https://api.hbdm.com/linear-swap-ex/market/detail/batch_merged"
	RefreshRate = 10 * time.Second
)

// ================= 数据结构 =================

// API 响应结构
type APIResponse struct {
	Success bool       `json:"success"`
	Data    DataResult `json:"data"`
}

type DataResult struct {
	Coins []CoinItem `json:"coins"`
	Count int        `json:"count"`
}

type AssetItem struct {
	Rank   int
	Symbol string
	Price  float64
	Volume float64 // 估算成交额
}

type CoinItem struct {
	Pair            string  `json:"pair"`
	Score           float64 `json:"score"`
	StartTime       int64   `json:"start_time"`
	StartPrice      float64 `json:"start_price"`
	LastScore       float64 `json:"last_score"`
	MaxScore        float64 `json:"max_score"`
	MaxPrice        float64 `json:"max_price"`
	IncreasePercent float64 `json:"increase_percent"`
}

// 全局缓存 (线程安全)
var (
	cache      APIResponse
	cacheMutex sync.RWMutex
)

// ================= 主程序 =================

func main() {
	setupLogging()

	// 1. 启动后台数据更新协程
	go backgroundFetcher()

	// 2. 设置 HTTP 路由
	http.HandleFunc("/api/ai500/list", handleAI500List)
	http.HandleFunc("/api/ai500/health", handleHealth)

	// 3. 启动服务器
	fmt.Printf("?? AI500 Service running at http://127.0.0.1%s/api/ai500/list\n", ServerPort)
	if err := http.ListenAndServe(ServerPort, nil); err != nil {
		log.Fatal(err)
	}
}

// ================= 处理逻辑 =================

// HTTP Handler
func handleAI500List(w http.ResponseWriter, r *http.Request) {
	// 允许跨域 (CORS) - 方便前端调用
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	cacheMutex.RLock()
	data := cache
	cacheMutex.RUnlock()

	// 如果缓存为空（服务刚启动），返回提示
	if !data.Success {
		http.Error(w, `{"error": "Data initializing..."}`, http.StatusServiceUnavailable)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// 健康检查：返回缓存状态，便于探活和监控
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cacheMutex.RLock()
	ready := cache.Success
	cacheMutex.RUnlock()

	status := http.StatusOK
	msg := "ok"
	if !ready {
		status = http.StatusServiceUnavailable
		msg = "initializing"
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": msg,
		"ready":  ready,
	})
}

// 后台循环抓取
func backgroundFetcher() {
	// 立即执行一次，然后定时
	updateData()
	ticker := time.NewTicker(RefreshRate)
	for range ticker.C {
		updateData()
	}
}

// 核心逻辑：从火币获取并计算
func updateData() {
	// 1. 获取火币所有行情
	ticks, err := fetchHuobiMarket()
	if err != nil {
		log.Printf("Error fetching Huobi data: %v", err)
		return
	}

	// 2. 计算所有返回的合约
	var items []AssetItem
	var totalIndex float64

	for _, t := range ticks {
		if !isPerpetual(t.ContractType) {
			continue
		}

		price, err := parseNumber(t.Close)
		if err != nil {
			log.Printf("skip %s: invalid close %q: %v", t.ContractCode, t.Close, err)
			continue
		}
		amount, err := parseNumber(t.Amount)
		if err != nil {
			log.Printf("skip %s: invalid amount %q: %v", t.ContractCode, t.Amount, err)
			continue
		}

		// 火币 amount 是币数，vol 是张数。估算 USD 成交额 = amount * price
		volUSD := amount * price

		// 过滤成交额过低的合约（以 USD 估算）
		if volUSD <= 10_000_000 {
			continue
		}

		items = append(items, AssetItem{
			Symbol: t.ContractCode,
			Price:  price,
			Volume: volUSD,
		})
		totalIndex += price
	}

	// 3. 排序 (按成交额降序)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Volume > items[j].Volume
	})

	// 4. 填充排名
	for i := range items {
		items[i].Rank = i + 1
	}

	// 5. 更新缓存
	coins := buildCoins(items)
	newResp := APIResponse{
		Success: true,
		Data: DataResult{
			Coins: coins,
			Count: len(coins),
		},
	}

	cacheMutex.Lock()
	cache = newResp
	cacheMutex.Unlock()

	log.Printf("Updated AI500 data. Index: %.2f, Count: %d", totalIndex, len(items))
}

// 初始化日志：输出到 stdout 和本地文件
func setupLogging() {
	logDir := "logs"
	writer, err := newRotatingWriter(logDir, "ai500-service", 30)
	if err != nil {
		log.Printf("创建日志文件失败，继续使用默认输出: %v", err)
		return
	}

	mw := io.MultiWriter(os.Stdout, writer)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("日志初始化完成，目录 %s，按天轮转，保留 30 天", logDir)
}

// 将内部资产结构转换为对外响应格式
func buildCoins(items []AssetItem) []CoinItem {
	if len(items) == 0 {
		return nil
	}

	maxVol := items[0].Volume
	start := time.Now().Add(-24 * time.Hour).Unix()

	coins := make([]CoinItem, 0, len(items))
	for _, it := range items {
		score := 0.0
		if maxVol > 0 {
			score = roundTo(it.Volume/maxVol*100, 1)
		}

		coins = append(coins, CoinItem{
			Pair:            strings.ReplaceAll(it.Symbol, "-", ""),
			Score:           score,
			StartTime:       start,
			StartPrice:      it.Price,
			LastScore:       score,
			MaxScore:        score,
			MaxPrice:        it.Price,
			IncreasePercent: 0,
		})
	}

	return coins
}

// 保留一定位数的小数，默认四舍五入
func roundTo(val float64, places int) float64 {
	if places <= 0 {
		return math.Round(val)
	}
	factor := math.Pow(10, float64(places))
	return math.Round(val*factor) / factor
}

// ================= 日志轮转工具 =================

type rotatingWriter struct {
	mu        sync.Mutex
	dir       string
	prefix    string
	retention int
	date      string
	file      *os.File
}

func newRotatingWriter(dir, prefix string, retention int) (io.Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	w := &rotatingWriter{
		dir:       dir,
		prefix:    prefix,
		retention: retention,
	}
	if err := w.rotateIfNeeded(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeeded(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

func (w *rotatingWriter) rotateIfNeeded() error {
	today := time.Now().Format("2006-01-02")
	if w.file != nil && w.date == today {
		return nil
	}

	if w.file != nil {
		_ = w.file.Close()
	}

	filename := fmt.Sprintf("%s-%s.log", w.prefix, today)
	path := filepath.Join(w.dir, filename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.date = today

	// 清理过期日志
	w.cleanupExpired()

	return nil
}

func (w *rotatingWriter) cleanupExpired() {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}

	re := regexp.MustCompile(fmt.Sprintf(`^%s-(\d{4}-\d{2}-\d{2})\.log$`, regexp.QuoteMeta(w.prefix)))
	cutoff := time.Now().AddDate(0, 0, -w.retention)

	for _, e := range entries {
		matches := re.FindStringSubmatch(e.Name())
		if len(matches) != 2 {
			continue
		}
		day, err := time.Parse("2006-01-02", matches[1])
		if err != nil {
			continue
		}
		if day.Before(cutoff) {
			_ = os.Remove(filepath.Join(w.dir, e.Name()))
		}
	}
}

// ================= 火币 API 工具 =================

type HuobiTicker struct {
	ContractCode string `json:"contract_code"`
	ContractType string `json:"contract_type"` // 永续标识，例如 swap
	Close        string `json:"close"`
	Amount       string `json:"amount"` // 成交量(币)
}

type HuobiResponse struct {
	Ticks []HuobiTicker `json:"ticks"`
}

func fetchHuobiMarket() ([]HuobiTicker, error) {
	// 检查是否跳过 TLS 验证 (仅用于 Docker 调试)
	tr := &http.Transport{}
	if os.Getenv("SKIP_TLS_VERIFY") == "true" {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := http.Client{Timeout: 10 * time.Second, Transport: tr}

	resp, err := client.Get(HuobiAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 只读请求失败时记录响应体，方便排查（截断防止过长）
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("Huobi status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, _ := io.ReadAll(resp.Body)
	var hResp HuobiResponse
	if err := json.Unmarshal(body, &hResp); err != nil {
		return nil, err
	}
	return hResp.Ticks, nil
}

// 将字符串数字转换为 float64，兼容接口返回字符串的情况
func parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// 判断是否永续合约；线性接口通常都是永续，但额外防御
func isPerpetual(contractType string) bool {
	if contractType == "" {
		return true
	}
	ct := strings.ToLower(contractType)
	return ct == "swap" || ct == "perpetual"
}
