package handler

import "encoding/xml"

// ---- CÁC STRUCT ĐỂ PARSE DỮ LIỆU ----

// Struct cho webhook từ Telegram
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ID int `json:"id"`
}

// Struct cho API giá Bitcoin (CoinGecko)
type BitcoinPrice struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
	} `json:"bitcoin"`
}

// Struct cho API tỷ giá (open.er-api.com)
type ExchangeRate struct {
	Rates map[string]float64 `json:"rates"`
}

// Struct cho XML giá vàng SJC
type SjcXML struct {
	XMLName  xml.Name `xml:"root"`
	Title    string   `xml:"title"`
	Url      string   `xml:"url"`
	Ratelist Ratelist `xml:"ratelist"`
}
type Ratelist struct {
	XMLName  xml.Name `xml:"ratelist"`
	City     []City   `xml:"city"`
	DateTime string   `xml:"updated"`
}
type City struct {
	XMLName xml.Name `xml:"city"`
	Name    string   `xml:"name,attr"`
	Item    []Item   `xml:"item"`
}
type Item struct {
	XMLName xml.Name `xml:"item"`
	Buy     string   `xml:"buy,attr"`
	Sell    string   `xml:"sell,attr"`
	Type    string   `xml:"type,attr"`
}

// Struct cho XML tỷ giá Vietcombank
type VcbExrateList struct {
	XMLName xml.Name    `xml:"ExrateList"`
	Exrate  []VcbExrate `xml:"Exrate"`
}
type VcbExrate struct {
	CurrencyCode string `xml:"CurrencyCode,attr"`
	CurrencyName string `xml:"CurrencyName,attr"`
	Buy          string `xml:"Buy,attr"`
	Transfer     string `xml:"Transfer,attr"`
	Sell         string `xml:"Sell,attr"`
}

// Struct cho APi vang.today (Full list)
type VangTodayResponse struct {
	Success   bool                `json:"success"`
	Timestamp int64               `json:"timestamp"`
	Prices    map[string]GoldItem `json:"prices"`
}

type GoldItem struct {
	Name       string  `json:"name"`
	Buy        float64 `json:"buy"`
	Sell       float64 `json:"sell"`
	ChangeBuy  float64 `json:"change_buy"`
	ChangeSell float64 `json:"change_sell"`
	Currency   string  `json:"currency"`
}

// Struct cho API vang.today (Single Type response, e.g. type=XAUUSD)
type VangTodaySingleResponse struct {
	Success    bool    `json:"success"`
	Timestamp  int64   `json:"timestamp"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Buy        float64 `json:"buy"`
	Sell       float64 `json:"sell"`
	ChangeBuy  float64 `json:"change_buy"`
	ChangeSell float64 `json:"change_sell"`
}

var goldTypeMap = map[string]string{
	"XAUUSD":      "Vàng Thế Giới (XAU/USD)",
	"SJL1L10":     "SJC 9999",
	"SJ9999":      "Nhẫn SJC",
	"DOHNL":       "DOJI Hà Nội",
	"DOHCML":      "DOJI HCM",
	"DOJINHTV":    "DOJI Nữ Trang",
	"BTSJC":       "Bảo Tín SJC",
	"BT9999NTT":   "Bảo Tín 9999",
	"PQHNVM":      "PNJ Hà Nội",
	"PQHN24NTT":   "PNJ 24K",
	"VNGSJC":      "VN Gold SJC",
	"VIETTINMSJC": "Viettin SJC",
}

var goldTypeOrder = []string{
	"XAUUSD",
	"SJL1L10",
	"SJ9999",
	"DOHNL",
	"DOHCML",
	"DOJINHTV",
	"BTSJC",
	"BT9999NTT",
	"PQHNVM",
	"PQHN24NTT",
	"VNGSJC",
	"VIETTINMSJC",
}
