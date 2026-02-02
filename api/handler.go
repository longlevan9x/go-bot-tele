package handler

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ---- CÁC STRUCT ĐỂ PARSE DỮ LIỆU ----

// Struct cho APi vang.today
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

// Struct cho API vang.today khi chỉ lấy 1 loại vàng
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

// ---- CÁC HÀM LẤY DỮ LIỆU ----

// ---- CÁC HÀM LẤY DỮ LIỆU ----

// ---- HÀM TIỆN ÍCH ĐỂ THỰC HIỆN YÊU CẦU HTTP ----
// Chúng ta cần hàm này vì Google sẽ chặn nếu không có User-Agent giống trình duyệt
func makeRequest(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Giả mạo User-Agent để yêu cầu trông giống như từ một trình duyệt thật
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	return client.Do(req)
}

// Lấy giá Bitcoin
func getBitcoinPrice() (string, error) {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var price BitcoinPrice
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return "", err
	}

	return fmt.Sprintf("💰 **Giá Bitcoin (USD):** `$%s`", formatFloat(price.Bitcoin.USD)), nil
}

// Lấy giá vàng thế giới (API vang.today)
func getGlobalGoldPrice() (string, error) {
	url := "https://www.vang.today/api/prices?type=XAUUSD"
	res, err := makeRequest(url)
	if err != nil {
		return "", fmt.Errorf("không thể truy cập vang.today: %v", err)
	}
	defer res.Body.Close()

	var data VangTodaySingleResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("lỗi đọc dữ liệu API: %v", err)
	}

	if !data.Success {
		return "", fmt.Errorf("API không trả về dữ liệu thành công")
	}

	return fmt.Sprintf("🥇 **Giá Vàng Thế Giới (USD/oz):** `$%s`", formatFloat(data.Buy)), nil
}

// Lấy giá vàng tổng hợp từ vang.today
func getVnGoldPrice() (string, error) {
	url := "https://www.vang.today/api/prices"

	res, err := makeRequest(url)
	if err != nil {
		return "", fmt.Errorf("không thể truy cập vang.today: %v", err)
	}
	defer res.Body.Close()

	var data VangTodayResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("lỗi đọc dữ liệu API: %v", err)
	}

	if !data.Success || len(data.Prices) == 0 {
		return "", fmt.Errorf("API không trả về dữ liệu thành công")
	}

	// data.Prices chính là map chúng ta cần
	dataMap := data.Prices

	// Format lại chuỗi kết quả
	var result strings.Builder
	result.WriteString("🏆 **Bảng Giá Vàng Tổng Hợp**\n")
	result.WriteString("------------------------------------\n")

	// Duyệt qua danh sách order để in theo thứ tự
	for _, code := range goldTypeOrder {
		item, exists := dataMap[code]
		if !exists {
			continue
		}

		name := goldTypeMap[code]

		// Xử lý hiển thị
		var buyStr, sellStr string

		if code == "XAUUSD" {
			buyStr = fmt.Sprintf("$%s", formatFloat(item.Buy))
			sellStr = fmt.Sprintf("$%s", formatFloat(item.Sell))
		} else {
			buyStr = fmt.Sprintf("%s VND", formatInt(int64(item.Buy)))
			sellStr = fmt.Sprintf("%s VND", formatInt(int64(item.Sell)))
		}

		result.WriteString(fmt.Sprintf("🔸 **%s**\n", name))
		result.WriteString(fmt.Sprintf("   • Mua: `%s`\n", buyStr))
		result.WriteString(fmt.Sprintf("   • Bán: `%s`\n", sellStr))
	}

	return result.String(), nil
}

// Lấy tỷ giá USD/JPY
func getUsdJpyRate() (string, error) {
	resp, err := http.Get("https://open.er-api.com/v6/latest/USD")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var rates ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return "", err
	}

	jpyRate := rates.Rates["JPY"]
	return fmt.Sprintf("🇺🇸/🇯🇵 **Tỷ giá USD/JPY:** `1 USD = %s JPY`", formatFloat(jpyRate)), nil
}

// Lấy tỷ giá JPY/VND từ Google Finance
func getJpyVndRate() (string, error) {
	url := "https://www.google.com/finance/quote/JPY-VND"

	res, err := makeRequest(url)
	if err != nil {
		return "", fmt.Errorf("không thể truy cập Google Finance: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("Google Finance trả về mã lỗi: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("lỗi đọc dữ liệu trang: %v", err)
	}

	// Đây là CSS Selector cho thẻ div chứa giá trị tỷ giá trên trang Google Finance
	// Selector này có thể thay đổi trong tương lai nếu Google cập nhật trang web
	priceStr := doc.Find(".YMlKec.fxKbKc").First().Text()

	if priceStr == "" {
		return "", fmt.Errorf("không tìm thấy tỷ giá trên trang Google Finance (có thể cấu trúc trang đã thay đổi)")
	}

	return fmt.Sprintf("🇯🇵/🇻🇳 **Tỷ giá JPY/VND (Google Finance):**\n`1 JPY = %s VND`", priceStr), nil
}

// ---- HÀM GỬI TIN NHẮN & HANDLER CHÍNH ----

// Hàm gửi tin nhắn về lại cho người dùng
func sendTelegramMessage(chatID int, text string) {
	// Lấy token từ biến môi trường mà Vercel cung cấp
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN environment variable not set")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)

	// Sử dụng struct hoặc map để marshal JSON chuẩn
	reqBody := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error sending message to Telegram: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err == nil {
			log.Printf("Telegram API Error: %v", respBody)
		} else {
			log.Printf("Telegram API Error (Status %d)", resp.StatusCode)
		}
	}
}

// Hàm handler chính mà Vercel sẽ gọi
func Handler(w http.ResponseWriter, r *http.Request) {
	// Kiểm tra xem có phải là cron job không
	if r.URL.Query().Get("mode") == "cron" {
		chatIDStr := os.Getenv("CHAT_ID")
		if chatIDStr == "" {
			log.Println("CHAT_ID not set for cron job")
			http.Error(w, "CHAT_ID not set", http.StatusInternalServerError)
			return
		}

		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			log.Printf("Invalid CHAT_ID: %v", err)
			http.Error(w, "Invalid CHAT_ID", http.StatusInternalServerError)
			return
		}

		// Lấy giá vàng
		goldPrice, err := getVnGoldPrice()
		if err != nil {
			log.Printf("Error getting gold price for cron: %v", err)
		} else {
			sendTelegramMessage(chatID, goldPrice)
		}

		// Lấy tỷ giá JPY/VND
		jpyRate, err := getJpyVndRate()
		if err != nil {
			log.Printf("Error getting JPY/VND rate for cron: %v", err)
		} else {
			sendTelegramMessage(chatID, jpyRate)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cron job executed"))
		return
	}

	// Xử lý webhook từ Telegram (POST request)
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		// Chỉ log nếu đây thực sự là POST request mà decode lỗi
		if r.Method == "POST" {
			log.Printf("Error decoding request: %v", err)
		}
		// Trả về 200 để Telegram không gửi lại request liên tục
		w.WriteHeader(http.StatusOK)
		return
	}

	if update.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var responseText string
	var err error

	// Phân tích lệnh từ người dùng
	log.Printf("Received message from Chat ID: %d", update.Message.Chat.ID)
	switch update.Message.Text {
	case "/start":
		responseText = "Chào mừng bạn đến với Bot Tra Cứu Giá! Hãy thử các lệnh: /bitcoin, /vang, /vangvn, /usdjpy, /jpyvnd"
	case "/bitcoin":
		responseText, err = getBitcoinPrice()
	case "/vang":
		responseText, err = getGlobalGoldPrice()
	case "/vangvn":
		responseText, err = getVnGoldPrice()
	case "/usdjpy":
		responseText, err = getUsdJpyRate()
	case "/jpyvnd":
		responseText, err = getJpyVndRate()
	default:
		responseText = "Lệnh không hợp lệ. Hãy thử /start để xem các lệnh có sẵn."
	}

	if err != nil {
		log.Printf("Error getting data for command %s: %v", update.Message.Text, err)
		responseText = fmt.Sprintf("Rất tiếc, đã có lỗi xảy ra khi lấy dữ liệu cho lệnh %s. Vui lòng thử lại sau.", update.Message.Text)
	}

	sendTelegramMessage(update.Message.Chat.ID, responseText)

	// Phản hồi lại cho Vercel là đã xử lý xong
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ---- HÀM TIỆN ÍCH ----
// Format số float cho dễ đọc
func formatFloat(num float64) string {
	s := strconv.FormatFloat(num, 'f', 2, 64)
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	result := ""
	for i, c := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result + "." + parts[1]
}

// Format số int có dấu phẩy ngăn cách hàng nghìn
func formatInt(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}
