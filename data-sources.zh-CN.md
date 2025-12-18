# 数据源API文档

> 本文档详细说明AI500币种池和OI Top持仓排行API的使用方法，包括请求参数、返回格式和配置示例。

---

## 📋 目录

- [AI500 币种池API](#ai500-币种池api)
- [OI Top 持仓排行API](#oi-top-持仓排行api)
- [在策略中配置](#在策略中配置)
- [调用示例](#调用示例)

---

## AI500 币种池API

### 接口概述

**用途**：获取基于AI评分排序的加密货币列表
**数据来源**：AI500评分系统
**更新频率**：实时

---

### 请求格式

```
GET {base_url}/api/ai500/list?auth={auth_key}
```

**完整示例**：
```
http://nofxaios.com:30006/api/ai500/list?auth=cm_568c67eae410d912c54c
```

---

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `auth` | string | ✅ 是 | 认证密钥 |

**注意**：此API不支持其他查询参数，返回完整币种列表。

---

### 返回数据结构

#### JSON格式

```json
{
  "success": true,
  "data": {
    "count": 150,
    "coins": [
      {
        "pair": "BTCUSDT",
        "score": 95.5,
        "last_score": 94.2,
        "max_score": 98.3,
        "start_time": 1734432000,
        "start_price": 95000.00,
        "max_price": 97000.00,
        "increase_percent": 2.5
      },
      {
        "pair": "ETHUSDT",
        "score": 88.3,
        "last_score": 87.5,
        "max_score": 92.1,
        "start_time": 1734432000,
        "start_price": 3500.00,
        "max_price": 3650.00,
        "increase_percent": 3.8
      },
      {
        "pair": "SOLUSDT",
        "score": 82.6,
        "last_score": 81.9,
        "max_score": 85.4,
        "start_time": 1734432000,
        "start_price": 95.50,
        "max_price": 98.20,
        "increase_percent": 1.8
      }
    ]
  }
}
```

---

### 字段说明

#### 顶层字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 请求是否成功 |
| `data` | object | 数据对象 |
| `data.count` | int | 返回的币种数量 |
| `data.coins` | array | 币种数据数组 |

#### coins数组字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `pair` | string | 交易对名称 | `"BTCUSDT"` |
| `score` | float | 当前AI评分（0-100） | `95.5` |
| `last_score` | float | 上一次评分 | `94.2` |
| `max_score` | float | 历史最高评分 | `98.3` |
| `start_time` | int64 | 统计开始时间（Unix时间戳） | `1734432000` |
| `start_price` | float | 开始价格（USDT） | `95000.00` |
| `max_price` | float | 统计期间最高价格 | `97000.00` |
| `increase_percent` | float | 价格涨幅百分比 | `2.5` |

---

### 评分说明

- **评分范围**：0-100
- **评分含义**：
  - `90-100`：极强信号，优先考虑
  - `80-90`：强信号，值得关注
  - `70-80`：中等信号
  - `< 70`：弱信号，谨慎参与

---

### 错误响应

#### 认证失败

```json
{
  "success": false,
  "error": "Invalid authentication key"
}
```

#### 服务器错误

```json
{
  "success": false,
  "error": "Internal server error"
}
```

---

## OI Top 持仓排行API

### 接口概述

**用途**：获取持仓量增加/减少排行榜
**数据来源**：币安交易所合约持仓数据
**更新频率**：2秒缓存（高频请求会命中缓存）

---

### 接口列表

| 接口路径 | 说明 |
|---------|------|
| `/api/oi/top-ranking` | 持仓量**增加**排行（支持自定义参数） |
| `/api/oi/low-ranking` | 持仓量**减少**排行（支持自定义参数） |
| `/api/oi/top` | 持仓增加Top20（固定参数，向后兼容） |

---

### 请求格式

#### 持仓增加排行

```
GET {base_url}/api/oi/top-ranking?limit={N}&duration={时间}&auth={auth_key}
```

**完整示例**：
```
http://nofxaios.com:30006/api/oi/top-ranking?limit=20&duration=1h&auth=cm_568c67eae410d912c54c
```

#### 持仓减少排行

```
GET {base_url}/api/oi/low-ranking?limit={N}&duration={时间}&auth={auth_key}
```

**完整示例**：
```
http://nofxaios.com:30006/api/oi/low-ranking?limit=20&duration=1h&auth=cm_568c67eae410d912c54c
```

---

### 请求参数

| 参数 | 类型 | 必填 | 默认值 | 取值范围 | 说明 |
|------|------|------|--------|---------|------|
| `limit` | int | ❌ 否 | `20` | 1-100 | 返回币种数量 |
| `duration` | string | ❌ 否 | `"1h"` | 见下表 | 统计时间范围 |
| `auth` | string | ✅ 是 | - | - | 认证密钥 |

---

### duration参数值

| 值 | 说明 | 推荐场景 |
|----|------|---------|
| `1m` | 1分钟 | 超短线交易 |
| `5m` | 5分钟 | 短线交易 |
| `15m` | 15分钟 | 日内交易 |
| `30m` | 30分钟 | 日内交易 |
| `1h` | 1小时（默认） | ⭐ **推荐**：日内波段 |
| `4h` | 4小时 | ⭐ **推荐**：波段交易 |
| `8h` | 8小时 | 波段交易 |
| `12h` | 12小时 | 中期趋势 |
| `24h` | 24小时 | ⭐ **推荐**：趋势交易 |
| `1d` | 1天（同24h） | 趋势交易 |
| `2d` | 2天 | 中长期趋势 |
| `3d` | 3天 | 中长期趋势 |

---

### 返回数据结构

#### JSON格式

```json
{
  "code": 0,
  "data": {
    "count": 20,
    "exchange": "binance",
    "time_range": "1小时",
    "time_range_param": "1h",
    "rank_type": "top",
    "limit": 20,
    "positions": [
      {
        "rank": 1,
        "symbol": "BTCUSDT",
        "oi_delta": 1500.5,
        "oi_delta_value": 145500000,
        "oi_delta_percent": 3.52,
        "current_oi": 44000,
        "price_delta_percent": 2.15,
        "net_long": 26000,
        "net_short": 18000
      },
      {
        "rank": 2,
        "symbol": "ETHUSDT",
        "oi_delta": 25000,
        "oi_delta_value": 87500000,
        "oi_delta_percent": 2.85,
        "current_oi": 900000,
        "price_delta_percent": 1.80,
        "net_long": 520000,
        "net_short": 380000
      }
    ]
  }
}
```

---

### 字段说明

#### 顶层字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 状态码（`0`=成功，非0=失败） |
| `data` | object | 数据对象 |

#### data字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `count` | int | 实际返回的币种数量 |
| `exchange` | string | 交易所名称（固定为`"binance"`） |
| `time_range` | string | 时间范围显示名称（如`"1小时"`） |
| `time_range_param` | string | 时间范围参数值（如`"1h"`） |
| `rank_type` | string | 排行类型：`"top"`增加 / `"low"`减少 |
| `limit` | int | 请求的数量限制 |
| `positions` | array | 持仓数据列表 |

#### positions数组字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `rank` | int | 排名（1为最高） | `1` |
| `symbol` | string | 币种交易对 | `"BTCUSDT"` |
| `oi_delta` | float | 持仓量变化（币的数量） | `1500.5` |
| `oi_delta_value` | float | **持仓价值变化（USDT）** ⭐ **排序依据** | `145500000` |
| `oi_delta_percent` | float | 持仓量变化百分比 | `3.52` |
| `current_oi` | float | 当前总持仓量（币的数量） | `44000` |
| `price_delta_percent` | float | 价格变化百分比 | `2.15` |
| `net_long` | float | 净多头持仓量（币的数量） | `26000` |
| `net_short` | float | 净空头持仓量（币的数量） | `18000` |

---

### 数据解读

#### 持仓量与价格的关系

| 持仓变化 | 价格变化 | 市场含义 | 交易策略 |
|---------|---------|---------|---------|
| ✅ 增加 | ⬆️ 上涨 | **多头主导** | 趋势可能延续，顺势做多 |
| ✅ 增加 | ⬇️ 下跌 | **空头主导** | 趋势可能延续，顺势做空 |
| ❌ 减少 | ⬆️ 上涨 | **空头平仓** | 可能是反弹，谨慎追多 |
| ❌ 减少 | ⬇️ 下跌 | **多头平仓** | 可能是回调，等待企稳 |

#### 多空比例判断

```
多空比 = net_long / net_short
```

- **多空比 > 1.5**：市场明显偏多，警惕反转风险
- **1.2 < 多空比 < 1.5**：市场偏多，正常多头趋势
- **0.8 < 多空比 < 1.2**：多空平衡
- **0.5 < 多空比 < 0.8**：市场偏空，正常空头趋势
- **多空比 < 0.5**：市场明显偏空，警惕反转风险

---

### 错误响应

| code | 说明 | 处理方式 |
|------|------|---------|
| `0` | 成功 | - |
| `401` | 认证失败（`auth`无效） | 检查认证密钥 |
| `400` | 参数错误 | 检查`limit`和`duration`参数 |
| `500` | 服务器内部错误 | 稍后重试或联系支持 |

---

## 在策略中配置

### AI500币种池配置

#### 方式1：使用AI500作为币种来源

```json
{
  "coin_source": {
    "source_type": "coinpool",
    "use_coin_pool": true,
    "coin_pool_limit": 10,
    "coin_pool_api_url": "http://nofxaios.com:30006/api/ai500/list?auth=YOUR_AUTH_KEY"
  }
}
```

**参数说明**：
- `source_type`: 设置为 `"coinpool"`
- `use_coin_pool`: 必须为 `true`
- `coin_pool_limit`: 取前N个评分最高的币种（建议5-15）
- `coin_pool_api_url`: **完整的API URL**（包含`?auth=xxx`参数）

**⚠️ 重要**：必须替换 `YOUR_AUTH_KEY` 为实际的认证密钥。

---

#### 方式2：混合模式（AI500 + 静态列表）

```json
{
  "coin_source": {
    "source_type": "mixed",
    "static_coins": ["BTCUSDT", "ETHUSDT"],
    "use_coin_pool": true,
    "coin_pool_limit": 8,
    "coin_pool_api_url": "http://nofxaios.com:30006/api/ai500/list?auth=YOUR_AUTH_KEY"
  }
}
```

**结果**：最终币种池 = `static_coins` + AI500前8个（自动去重）

---

### OI Top配置

#### 方式1：OI排行数据（推荐）

用于AI分析市场持仓量变化趋势。

```json
{
  "indicators": {
    "enable_oi_ranking": true,
    "oi_ranking_api_url": "http://nofxaios.com:30006",
    "oi_ranking_duration": "1h",
    "oi_ranking_limit": 10
  }
}
```

**参数说明**：
- `enable_oi_ranking`: 是否启用OI排行数据
- `oi_ranking_api_url`: API基础URL（**仅基础URL，不含路径**）
- `oi_ranking_duration`: 时间范围（`"1h"`, `"4h"`, `"24h"`）
- `oi_ranking_limit`: 获取数量（1-100）

**⚠️ 重要**：系统会自动拼接完整路径：
- Top增加：`{base_url}/api/oi/top-ranking?limit={N}&duration={时间}&auth={key}`
- Low减少：`{base_url}/api/oi/low-ranking?limit={N}&duration={时间}&auth={key}`

---

#### 方式2：OI Top作为币种来源（已弃用）

不推荐使用，建议改用方式1。

```json
{
  "coin_source": {
    "source_type": "oi_top",
    "use_oi_top": true,
    "oi_top_limit": 20,
    "oi_top_api_url": "http://nofxaios.com:30006/api/oi/top-ranking?limit=20&duration=1h&auth=YOUR_AUTH_KEY"
  }
}
```

---

### 完整配置示例

结合AI500币种池和OI排行数据：

```json
{
  "coin_source": {
    "source_type": "coinpool",
    "use_coin_pool": true,
    "coin_pool_limit": 10,
    "coin_pool_api_url": "http://nofxaios.com:30006/api/ai500/list?auth=cm_568c67eae410d912c54c"
  },
  "indicators": {
    "klines": {
      "primary_timeframe": "5m",
      "primary_count": 30,
      "selected_timeframes": ["5m", "15m", "1h", "4h"]
    },
    "enable_cvd": true,
    "enable_vwap": true,
    "enable_oi": true,
    "enable_oi_ranking": true,
    "oi_ranking_api_url": "http://nofxaios.com:30006",
    "oi_ranking_duration": "1h",
    "oi_ranking_limit": 10
  }
}
```

---

## 调用示例

### Python

```python
import requests

# ========== AI500 API ==========
url = "http://nofxaios.com:30006/api/ai500/list"
params = {"auth": "cm_568c67eae410d912c54c"}
response = requests.get(url, params=params)
data = response.json()

if data["success"]:
    coins = data["data"]["coins"]
    print(f"✓ AI500返回 {len(coins)} 个币种")
    for coin in coins[:5]:  # 显示前5个
        print(f"  {coin['pair']}: 评分={coin['score']}, 涨幅={coin['increase_percent']}%")
else:
    print(f"✗ 请求失败: {data.get('error', 'Unknown error')}")

# ========== OI Top API ==========
url = "http://nofxaios.com:30006/api/oi/top-ranking"
params = {
    "limit": 20,
    "duration": "1h",
    "auth": "cm_568c67eae410d912c54c"
}
response = requests.get(url, params=params)
data = response.json()

if data["code"] == 0:
    positions = data["data"]["positions"]
    print(f"\n✓ OI Top返回 {len(positions)} 个币种 (时间范围: {data['data']['time_range']})")
    for pos in positions[:5]:  # 显示前5个
        oi_value = pos['oi_delta_value']
        oi_percent = pos['oi_delta_percent']
        price_percent = pos['price_delta_percent']
        print(f"  #{pos['rank']} {pos['symbol']}: "
              f"OI变化=${oi_value:,.0f} ({oi_percent:+.2f}%), "
              f"价格{price_percent:+.2f}%")
else:
    print(f"✗ 请求失败: code={data['code']}")
```

---

### JavaScript / Node.js

```javascript
const axios = require('axios');

// ========== AI500 API ==========
async function fetchAI500() {
  const url = 'http://nofxaios.com:30006/api/ai500/list';
  const params = { auth: 'cm_568c67eae410d912c54c' };

  try {
    const response = await axios.get(url, { params });
    const data = response.data;

    if (data.success) {
      const coins = data.data.coins;
      console.log(`✓ AI500返回 ${coins.length} 个币种`);
      coins.slice(0, 5).forEach(coin => {
        console.log(`  ${coin.pair}: 评分=${coin.score}, 涨幅=${coin.increase_percent}%`);
      });
    } else {
      console.error(`✗ 请求失败: ${data.error || 'Unknown error'}`);
    }
  } catch (error) {
    console.error(`✗ 请求异常: ${error.message}`);
  }
}

// ========== OI Top API ==========
async function fetchOITop() {
  const url = 'http://nofxaios.com:30006/api/oi/top-ranking';
  const params = {
    limit: 20,
    duration: '1h',
    auth: 'cm_568c67eae410d912c54c'
  };

  try {
    const response = await axios.get(url, { params });
    const data = response.data;

    if (data.code === 0) {
      const positions = data.data.positions;
      console.log(`\n✓ OI Top返回 ${positions.length} 个币种 (时间范围: ${data.data.time_range})`);
      positions.slice(0, 5).forEach(pos => {
        console.log(`  #${pos.rank} ${pos.symbol}: ` +
          `OI变化=$${pos.oi_delta_value.toLocaleString()} (${pos.oi_delta_percent > 0 ? '+' : ''}${pos.oi_delta_percent.toFixed(2)}%), ` +
          `价格${pos.price_delta_percent > 0 ? '+' : ''}${pos.price_delta_percent.toFixed(2)}%`);
      });
    } else {
      console.error(`✗ 请求失败: code=${data.code}`);
    }
  } catch (error) {
    console.error(`✗ 请求异常: ${error.message}`);
  }
}

// 执行
fetchAI500();
fetchOITop();
```

---

### cURL

```bash
# ========== AI500 API ==========
curl -X GET "http://nofxaios.com:30006/api/ai500/list?auth=cm_568c67eae410d912c54c"

# ========== OI Top API ==========
curl -X GET "http://nofxaios.com:30006/api/oi/top-ranking?limit=20&duration=1h&auth=cm_568c67eae410d912c54c"

# ========== OI Low API ==========
curl -X GET "http://nofxaios.com:30006/api/oi/low-ranking?limit=20&duration=1h&auth=cm_568c67eae410d912c54c"
```

---

### Go (项目内部实现参考)

```go
import (
    "nofx/provider"
    "log"
)

// 获取AI500数据
func getAI500Data() {
    provider.SetAI500API("http://nofxaios.com:30006/api/ai500/list?auth=cm_568c67eae410d912c54c")

    coins, err := provider.GetTopRatedCoins(10)
    if err != nil {
        log.Printf("❌ Failed to get AI500 data: %v", err)
        return
    }

    log.Printf("✓ AI500 top 10 coins: %v", coins)
}

// 获取OI Top数据
func getOITopData() {
    provider.SetOITopAPI("http://nofxaios.com:30006/api/oi/top-ranking?limit=20&duration=1h&auth=cm_568c67eae410d912c54c")

    positions, err := provider.GetOITopPositions()
    if err != nil {
        log.Printf("❌ Failed to get OI Top data: %v", err)
        return
    }

    log.Printf("✓ OI Top positions: %d coins", len(positions))
    for _, pos := range positions[:5] {
        log.Printf("  #%d %s: OI变化=$%.0f (%.2f%%), 价格%.2f%%",
            pos.Rank, pos.Symbol, pos.OIDeltaValue, pos.OIDeltaPercent, pos.PriceDeltaPercent)
    }
}
```

---

## 注意事项

### 通用

1. **认证密钥**：示例中的 `cm_568c67eae410d912c54c` 为演示密钥，实际使用时需替换为真实密钥
2. **速率限制**：OI Top API有2秒缓存，避免高频无意义请求
3. **HTTPS支持**：生产环境建议使用HTTPS（如API支持）
4. **错误处理**：务必检查返回值中的 `success`/`code` 字段
5. **超时设置**：建议设置30秒请求超时

---

### AI500特定

1. **评分时效性**：评分实时更新，建议每3-5分钟刷新一次
2. **币种数量**：通常返回100+个币种，根据需要使用 `coin_pool_limit` 限制
3. **过滤逻辑**：系统会自动过滤不可交易的币种

---

### OI Top特定

1. **排序依据**：按 `oi_delta_value`（持仓价值变化USDT）排序，而非持仓量变化
2. **数据来源**：仅支持币安交易所数据
3. **时间范围**：`duration` 参数影响数据的时效性，短周期适合短线，长周期适合趋势
4. **limit限制**：最大值为100，超过会被截断

---

## 故障排查

### 问题1：401认证失败

**症状**：返回 `"Invalid authentication key"` 或 `code: 401`

**解决方案**：
1. 检查URL中的 `auth` 参数是否正确
2. 确认密钥没有过期
3. 确认URL中没有多余的空格或特殊字符

---

### 问题2：返回数据为空

**症状**：`data.coins` 或 `data.positions` 为空数组

**可能原因**：
- AI500：当前无可交易币种（罕见）
- OI Top：指定时间范围内无持仓变化数据

**解决方案**：
- 检查API服务状态
- 尝试更换 `duration` 参数（OI Top）
- 联系API提供方确认

---

### 问题3：请求超时

**症状**：请求耗时超过30秒或连接超时

**解决方案**：
1. 检查网络连接
2. 确认API服务器地址正确
3. 尝试更换DNS或网络环境
4. 联系API提供方确认服务状态

---

### 问题4：策略配置后数据未生效

**症状**：修改策略配置后，AI决策中仍无OI排行数据

**解决方案**：
1. 确认策略配置已保存
2. **重启Trader**（配置只在启动时加载）
3. 检查日志中是否有 `"OI ranking data ready"` 信息
4. 确认 `oi_ranking_api_url` 只填基础URL，不含路径

---

## 相关文档

- [策略模块架构](../architecture/STRATEGY_MODULE.zh-CN.md)
- [OI API详细文档](./oi_api.md)
- [故障排查指南](../guides/TROUBLESHOOTING.zh-CN.md)

---

## 版本历史

| 版本 | 日期 | 变更说明 |
|------|------|---------|
| 1.0 | 2024-12-18 | 初始版本，整合AI500和OI Top API文档 |

---

**文档维护**: NoFx开发团队
**最后更新**: 2024-12-18
**API版本**: v1.0
