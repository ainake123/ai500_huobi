package main

import (
	"fmt"
	"os"
	"time"
)

func testPersistence() {
	fmt.Println("==========================================")
	fmt.Println("持久化单元测试")
	fmt.Println("==========================================\n")

	// 1. 清理旧数据
	fmt.Println("1️⃣  清理旧数据...")
	os.RemoveAll("data")

	// 2. 创建测试数据
	fmt.Println("\n2️⃣  创建测试历史数据...")
	historyMutex.Lock()
	historyCache["BTCUSDT"] = &CoinHistory{
		FirstSeen:  time.Now(),
		StartPrice: 50000.0,
		MaxPrice:   52000.0,
		LastScore:  95.0,
		MaxScore:   98.5,
	}
	historyCache["ETHUSDT"] = &CoinHistory{
		FirstSeen:  time.Now(),
		StartPrice: 3000.0,
		MaxPrice:   3200.0,
		LastScore:  88.0,
		MaxScore:   92.0,
	}
	historyMutex.Unlock()

	fmt.Printf("   创建了 %d 个测试币种\n", len(historyCache))

	// 3. 保存到文件
	fmt.Println("\n3️⃣  保存历史数据到文件...")
	if err := saveHistoryToFile(); err != nil {
		fmt.Printf("   ❌ 保存失败: %v\n", err)
		return
	}
	fmt.Println("   ✅ 保存成功")

	// 4. 检查文件
	fmt.Println("\n4️⃣  检查历史文件...")
	if info, err := os.Stat(HistoryFilePath); err == nil {
		fmt.Printf("   ✅ 文件已生成: %s\n", HistoryFilePath)
		fmt.Printf("   文件大小: %d 字节\n", info.Size())
	} else {
		fmt.Printf("   ❌ 文件不存在: %v\n", err)
		return
	}

	// 5. 清空内存缓存
	fmt.Println("\n5️⃣  清空内存缓存...")
	historyMutex.Lock()
	originalCount := len(historyCache)
	historyCache = make(map[string]*CoinHistory)
	historyMutex.Unlock()
	fmt.Printf("   内存中币种数量: %d -> %d\n", originalCount, len(historyCache))

	// 6. 从文件加载
	fmt.Println("\n6️⃣  从文件加载历史数据...")
	if err := loadHistoryFromFile(); err != nil {
		fmt.Printf("   ❌ 加载失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 加载成功，共 %d 个币种\n", len(historyCache))

	// 7. 验证数据完整性
	fmt.Println("\n7️⃣  验证数据完整性...")
	success := true

	if btc, exists := historyCache["BTCUSDT"]; exists {
		fmt.Println("   BTC 数据:")
		fmt.Printf("     StartPrice: %.2f\n", btc.StartPrice)
		fmt.Printf("     MaxPrice: %.2f\n", btc.MaxPrice)
		fmt.Printf("     LastScore: %.1f\n", btc.LastScore)
		fmt.Printf("     MaxScore: %.1f\n", btc.MaxScore)

		if btc.StartPrice != 50000.0 || btc.MaxPrice != 52000.0 {
			fmt.Println("   ❌ BTC 价格数据不匹配")
			success = false
		} else if btc.LastScore != 95.0 || btc.MaxScore != 98.5 {
			fmt.Println("   ❌ BTC 评分数据不匹配")
			success = false
		} else {
			fmt.Println("   ✅ BTC 数据完整")
		}
	} else {
		fmt.Println("   ❌ BTC 数据丢失")
		success = false
	}

	if eth, exists := historyCache["ETHUSDT"]; exists {
		fmt.Println("\n   ETH 数据:")
		fmt.Printf("     StartPrice: %.2f\n", eth.StartPrice)
		fmt.Printf("     MaxPrice: %.2f\n", eth.MaxPrice)
		fmt.Printf("     LastScore: %.1f\n", eth.LastScore)
		fmt.Printf("     MaxScore: %.1f\n", eth.MaxScore)

		if eth.StartPrice != 3000.0 || eth.MaxPrice != 3200.0 {
			fmt.Println("   ❌ ETH 价格数据不匹配")
			success = false
		} else if eth.LastScore != 88.0 || eth.MaxScore != 92.0 {
			fmt.Println("   ❌ ETH 评分数据不匹配")
			success = false
		} else {
			fmt.Println("   ✅ ETH 数据完整")
		}
	} else {
		fmt.Println("   ❌ ETH 数据丢失")
		success = false
	}

	fmt.Println("\n==========================================")
	if success {
		fmt.Println("✅ 持久化测试全部通过！")
	} else {
		fmt.Println("❌ 持久化测试失败！")
	}
	fmt.Println("==========================================")
}

func main() {
	testPersistence()
}
