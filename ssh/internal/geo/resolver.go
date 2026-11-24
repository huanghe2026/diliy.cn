// Package geo 轻量 IP → 地理位置 resolver
// 多源 fallback + 内存缓存 + 2s 超时
package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Location IP 地理信息
type Location struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Org         string  `json:"org"` // ISP / 组织
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
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
//  格式: { city, region, country, org, loc: "lat,lon", ... }
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

	lat, lon := parseLoc(strField(raw, "loc"))

	return &Location{
		City:        strField(raw, "city"),
		Region:      strField(raw, "region"),
		CountryCode: strField(raw, "country"),
		CountryName: countryName(raw["country"]),
		Org:         strField(raw, "org"),
		Lat:         lat,
		Lon:         lon,
	}
}

// ═══════════════════════════════════════════════
//  📡 源 2: ip-api.com (免费, 45 req/min, 无 key)
//  格式: { status, lat, lon, countryCode, regionName, city, isp, org, ... }
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

	var lat, lon float64
	if v, ok := raw["lat"].(float64); ok {
		lat = v
	}
	if v, ok := raw["lon"].(float64); ok {
		lon = v
	}

	return &Location{
		City:        strField(raw, "city"),
		Region:      strField(raw, "regionName"),
		CountryCode: strField(raw, "countryCode"),
		CountryName: strField(raw, "country"),
		Org:         firstNonEmpty(strField(raw, "org"), strField(raw, "isp")),
		Lat:         lat,
		Lon:         lon,
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

// parseLoc ipinfo loc 格式 "lat,lon" → 两个 float64
func parseLoc(loc string) (float64, float64) {
	parts := strings.Split(loc, ",")
	if len(parts) != 2 {
		return 0, 0
	}
	lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lon, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return lat, lon
}

// LatLonFromIP — 便利函数：直接拿某 IP 的经纬度，拿不到返回 (0, 0)
func LatLonFromIP(ip string) (float64, float64) {
	l := Resolve(ip)
	if l == nil || (l.Lat == 0 && l.Lon == 0) {
		return 0, 0
	}
	return l.Lat, l.Lon
}

// fmt 是 import 了但没用到 — 去掉未使用？加个简单 wrapper
var _ = fmt.Sprintf

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
