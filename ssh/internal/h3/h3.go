// Package h3 — H3 六边形网格编码（Geoyuan 地理元核心）
// Uber H3 v4 纯 Go 库：经纬度 → H3 cell → k-ring/邻居/距离/面积
//
// Geoyuan 用法：
//   mint 卡片 → lat/lon → H3 cell (uint64) 存链上
//   附近卡片 → GridDisk(cell, k=2) → 19 cells → 查链上
//   邻居配对 → GridRing(cell, 1) → 6 邻居
package h3

import (
	"fmt"
	"math"

	h3lib "github.com/uber/h3-go/v4"
)

// DefaultRes — Geoyuan 默认精度 9（~7 km²/hex，街区级）
const DefaultRes = 9

// Cell — Geoyuan 地理元单元格
type Cell struct {
	Index   uint64  // H3 Cell (int64 底层)
	String  string  // hex 字符串 — 链上 tokenURI
	Lat     float64 // 中心纬度
	Lon     float64 // 中心经度
	Res     int     // 精度 level (0-15)
	AreaKm2 float64 // 面积
}

// Encode — 经纬度 → H3 Cell
func Encode(lat, lon float64, res int) Cell {
	if res <= 0 {
		res = DefaultRes
	}
	ll := h3lib.NewLatLng(lat, lon)
	idx, _ := h3lib.LatLngToCell(ll, res)
	return cellFromIdx(idx)
}

// Decode — uint64 → Cell（反查中心坐标 + 面积）
func Decode(index uint64) Cell {
	return cellFromIdx(h3lib.Cell(index))
}

// cellFromIdx — h3lib.Cell → Cell
func cellFromIdx(idx h3lib.Cell) Cell {
	center, _ := h3lib.CellToLatLng(idx)
	area, _ := h3lib.CellAreaKm2(idx)
	res := extractRes(uint64(idx))
	return Cell{
		Index:   uint64(idx),
		String:  fmt.Sprintf("%x", uint64(idx)),
		Lat:     center.Lat,
		Lon:     center.Lng,
		Res:     res,
		AreaKm2: area,
	}
}

// extractRes — 从 H3 uint64 index 提取 resolution (0-15)
// H3 index 位布局：MSB 保留位(4) + res(4) + baseCell(7) + 子位 digit*7
func extractRes(idx uint64) int {
	return int((idx >> 52) & 0x0F)
}

// Neighbors — 6 个相邻 cell（GridRing k=1 去掉自己）
func Neighbors(index uint64) []Cell {
	ring, err := h3lib.Cell(h3lib.Cell(index)).GridRing(1)
	if err != nil {
		return nil
	}
	out := make([]Cell, 0, len(ring))
	for _, n := range ring {
		out = append(out, Decode(uint64(n)))
	}
	return out
}

// KRing — 自己 + k 层邻居
// k=0 → 1 cell, k=1 → 7 cells, k=2 → 19 cells, k=3 → 37 cells
func KRing(index uint64, k int) []Cell {
	disk, err := h3lib.Cell(h3lib.Cell(index)).GridDisk(k)
	if err != nil {
		return nil
	}
	out := make([]Cell, 0, len(disk))
	for _, i := range disk {
		out = append(out, Decode(uint64(i)))
	}
	return out
}

// GridDistance — 两个 cell 的 H3 网格跳跃数（快！不涉及球面计算）
func GridDistance(idx1, idx2 uint64) int {
	d, _ := h3lib.GridDistance(h3lib.Cell(idx1), h3lib.Cell(idx2))
	return d
}

// DistanceKm — 球面距离 km
func DistanceKm(idx1, idx2 uint64) float64 {
	c1 := Decode(idx1)
	c2 := Decode(idx2)
	return haversineKm(c1.Lat, c1.Lon, c2.Lat, c2.Lon)
}

// GreatCircle — h3 内置更快的大圆距离
func GreatCircle(idx1, idx2 uint64) float64 {
	c1 := Decode(idx1)
	c2 := Decode(idx2)
	a := h3lib.NewLatLng(c1.Lat, c1.Lon)
	b := h3lib.NewLatLng(c2.Lat, c2.Lon)
	return h3lib.GreatCircleDistanceKm(a, b)
}

// AreaKm2 — uint64 index → 面积 km²（Uber 内置 CellAreaKm2 wrapper）
func AreaKm2(idx uint64) float64 {
	a, _ := h3lib.CellAreaKm2(h3lib.Cell(idx))
	return a
}

// BoundingBoxCells — 矩形 (lat1,lon1)×(lat2,lon2) 覆盖的所有 H3 cells
func BoundingBoxCells(lat1, lon1, lat2, lon2 float64, res int) []Cell {
	poly := h3lib.GeoPolygon{
		GeoLoop: h3lib.GeoLoop{
			h3lib.NewLatLng(lat1, lon1),
			h3lib.NewLatLng(lat2, lon1),
			h3lib.NewLatLng(lat2, lon2),
			h3lib.NewLatLng(lat1, lon2),
		},
	}
	cells, err := poly.Cells(res)
	if err != nil {
		return nil
	}
	out := make([]Cell, 0, len(cells))
	for _, i := range cells {
		out = append(out, Decode(uint64(i)))
	}
	return out
}

// CellBoundary — H3 cell 六边形顶点坐标（可用来画 hex 地图）
func CellBoundary(index uint64) [][2]float64 {
	b, err := h3lib.CellToBoundary(h3lib.Cell(index))
	if err != nil {
		return nil
	}
	out := make([][2]float64, len(b))
	for i, p := range b {
		out[i] = [2]float64{p.Lat, p.Lng}
	}
	return out
}

// haversine — 球面距离
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	R := 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
