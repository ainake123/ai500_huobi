package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	Timestamp int64       `json:"timestamp"`
	Index     float64     `json:"index_value"`
	Count     int         `json:"count"`
	Data      []AssetItem `json:"list"`
}

type AssetItem struct {
	Rank   int     `json:"rank"`
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Volume float64 `json:"volume_24h_usd"` // 估算成交额
}

// 全局缓存 (线程安全)
var (
	cache      APIResponse
	cacheMutex sync.RWMutex
)

// ================= 主程序 =================

func main() {
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
	if data.Timestamp == 0 {
		http.Error(w, `{"error": "Data initializing..."}`, http.StatusServiceUnavailable)
		return
	}

	json.NewEncoder(w).Encode(data)
}

// 健康检查：返回缓存状态，便于探活和监控
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cacheMutex.RLock()
	ready := cache.Timestamp != 0
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
	newResp := APIResponse{
		Timestamp: time.Now().Unix(),
		Index:     totalIndex,
		Count:     len(items),
		Data:      items,
	}

	cacheMutex.Lock()
	cache = newResp
	cacheMutex.Unlock()
	
	log.Printf("Updated AI500 data. Index: %.2f, Count: %d", totalIndex, len(items))
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
	client := http.Client{Timeout: 5 * time.Second}
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
