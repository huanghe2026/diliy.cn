// Package geo 轻量 IP → 地理位置 resolver
// 用 ipapi.co 免费 API，带内存缓存 + 2s 超时
package geo

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Location IP 地理信息（精简版）
type Location struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Region      string `json:"region"`
	City        string `json:"city"`
	Org         string `json:"org"` // ISP / 组织

	// Raw API response (ipapi.co)
	raw map[string]interface{}
}

// Display 生成人类可读的位置字符串
// 优先返回 "City, Country"，降级到 "Country"，都没有返回 ""
func (l *Location) Display() string {
	if l == nil {
		return ""
	}
	if l.City != "" && l.CountryName != "" {
		return l.City + ", " + l.CountryName
	}
	if l.City != "" {
		return l.City
	}
	if l.CountryName != "" {
		return l.CountryName
	}
	return ""
}

// Full 完整显示（带 ISP）
func (l *Location) Full() string {
	if l == nil {
		return ""
	}
	loc := l.Display()
	if l.Org != "" && loc != "" {
		return loc + " · " + l.Org
	}
	if l.Org != "" {
		return l.Org
	}
	return loc
}

// ═══════════════════════════════════════════════
//  🔌 Resolver — 带缓存的 IP 地理位置查询
// ═══════════════════════════════════════════════

// cacheEntry 带过期时间的缓存项
type cacheEntry struct {
	loc *Location
	exp time.Time
}

const cacheTTL = 24 * time.Hour // 同一 IP 缓存 24h

var (
	cacheMu sync.RWMutex
	cache   = make(map[string]cacheEntry)
	httpCli = &http.Client{Timeout: 2 * time.Second}
)

// Resolve 查一个 IP 的地理位置（带缓存）
// 失败返回 nil, "" — 调用方降级显示 IP
func Resolve(ip string) *Location {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return nil
	}

	// 查缓存
	cacheMu.RLock()
	if e, ok := cache[ip]; ok && time.Now().Before(e.exp) {
		cacheMu.RUnlock()
		return e.loc
	}
	cacheMu.RUnlock()

	// 查 API
	loc := queryIPAPI(ip)

	// 写缓存（哪怕 nil — 避免同一 IP 反复打 API）
	cacheMu.Lock()
	cache[ip] = cacheEntry{loc: loc, exp: time.Now().Add(cacheTTL)}
	cacheMu.Unlock()

	return loc
}

// queryIPAPI 实际查 ipapi.co
func queryIPAPI(ip string) *Location {
	url := "https://ipapi.co/" + ip + "/json/"
	resp, err := httpCli.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil
	}

	// ipapi.co 对私有 IP 返回 "Private IP Address"
	if v, _ := raw["error"].(bool); v {
		return nil
	}

	loc := &Location{
		CountryCode: strField(raw, "country_code"),
		CountryName: strField(raw, "country_name"),
		Region:      strField(raw, "region"),
		City:        strField(raw, "city"),
		Org:         strField(raw, "org"),
		raw:         raw,
	}
	return loc
}

// strField 安全取 string 字段
func strField(m map[string]interface{}, key string) string {
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return v
}

// ═══════════════════════════════════════════════
//  🔄 备用 API — 多个 source 增加鲁棒性
// ═══════════════════════════════════════════════
// ipapi.co 限免费 1000 req/day，够用
// 如果挂了可以切换到 ip.sb/geoip/ip

// ClearCache 清缓存（主要用于测试）
func ClearCache() {
	cacheMu.Lock()
	cache = make(map[string]cacheEntry)
	cacheMu.Unlock()
}
