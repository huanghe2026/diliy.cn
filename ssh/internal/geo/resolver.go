// Package geo 轻量 IP → 地理位置 resolver
// 多源 fallback + 内存缓存 + 2s 超时
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

// ═══════════════════════════════════════════════
//  🔌 Resolver — 带缓存的 IP 地理位置查询
// ═══════════════════════════════════════════════

// cacheEntry 带过期时间的缓存项
type cacheEntry struct {
	loc *Location
	exp time.Time
}

const cacheTTL = 24 * time.Hour

var (
	cacheMu sync.RWMutex
	cache   = make(map[string]cacheEntry)
	httpCli = &http.Client{Timeout: 2 * time.Second}
)

// Resolve 查一个 IP 的地理位置（带缓存 + 多源 fallback）
// 失败返回 nil — 调用方降级显示 IP
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

	// 多源 fallback：ipinfo.io → ip-api.com
	loc := queryIPinfo(ip)
	if loc == nil {
		loc = queryIPAPI(ip)
	}

	// 写缓存
	cacheMu.Lock()
	cache[ip] = cacheEntry{loc: loc, exp: time.Now().Add(cacheTTL)}
	cacheMu.Unlock()

	return loc
}

// ═══════════════════════════════════════════════
//  📡 源 1: ipinfo.io (免费 50k req/月)
//  格式: { city, region, country, org, ... }
// ═══════════════════════════════════════════════
func queryIPinfo(ip string) *Location {
	resp, err := httpCli.Get("https://ipinfo.io/" + ip + "/json")
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

	// 这个源对私有 IP 也会返回有效数据，不需要 error 检查
	return &Location{
		City:        strField(raw, "city"),
		Region:      strField(raw, "region"),
		CountryCode: strField(raw, "country"),
		CountryName: countryName(raw["country"]),
		Org:         strField(raw, "org"),
	}
}

// ═══════════════════════════════════════════════
//  📡 源 2: ip-api.com (免费, 45 req/min, 无 key)
//  格式: { status, country, countryCode, regionName, city, isp, org, ... }
// ═══════════════════════════════════════════════
func queryIPAPI(ip string) *Location {
	resp, err := httpCli.Get("http://ip-api.com/json/" + ip)
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
	if strField(raw, "status") != "success" {
		return nil
	}

	return &Location{
		City:        strField(raw, "city"),
		Region:      strField(raw, "regionName"),
		CountryCode: strField(raw, "countryCode"),
		CountryName: strField(raw, "country"),
		Org:         firstNonEmpty(strField(raw, "org"), strField(raw, "isp")),
	}
}

// ═══════════════════════════════════════════════
//  🔧 辅助函数
// ═══════════════════════════════════════════════

func strField(m map[string]interface{}, key string) string {
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return v
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

// countryName 把 ISO code 转全名（简单 map，够用）
var countryNames = map[string]string{
	"CN": "China", "US": "United States", "JP": "Japan", "KR": "South Korea",
	"SG": "Singapore", "HK": "Hong Kong", "TW": "Taiwan", "GB": "United Kingdom",
	"DE": "Germany", "FR": "France", "NL": "Netherlands", "RU": "Russia",
	"IN": "India", "BR": "Brazil", "AU": "Australia", "CA": "Canada",
}

func countryName(code interface{}) string {
	if code == nil {
		return ""
	}
	c, ok := code.(string)
	if !ok {
		return ""
	}
	if name, ok := countryNames[c]; ok {
		return name
	}
	return c // 没映射到就返回 code 本身
}

// ClearCache 清缓存（主要用于测试）
func ClearCache() {
	cacheMu.Lock()
	cache = make(map[string]cacheEntry)
	cacheMu.Unlock()
}
