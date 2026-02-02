package handler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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


// ---- CÁC HÀM LẤY DỮ LIỆU ----

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

// Lấy giá vàng thế giới (Scraping)
func getGlobalGoldPrice() (string, error) {
	res, err := http.Get("https://goldprice.org/")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	priceStr := doc.Find("#gpxticker-spot-bid").Text()
	priceStr = strings.Replace(priceStr, ",", "", -1) // Bỏ dấu phẩy
	price, _ := strconv.ParseFloat(priceStr, 64)

	return fmt.Sprintf("🥇 **Giá Vàng Thế Giới (USD/oz):** `$%s`", formatFloat(price)), nil
}

// Lấy giá vàng SJC
func getVnGoldPrice() (string, error) {
	// SJC cung cấp file XML công khai
	resp, err := http.Get("https://sjc.com.vn/xml/tygiavang.xml")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	byteValue, _ := io.ReadAll(resp.Body)
	var sjcData SjcXML
	xml.Unmarshal(byteValue, &sjcData)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("🇻🇳 **Giá Vàng SJC** (cập nhật lúc %s)\n", sjcData.Ratelist.DateTime))
	result.WriteString("------------------------------------\n")

	// Lấy giá tại TP.HCM
	for _, city := range sjcData.Ratelist.City {
		if city.Name == "Hồ Chí Minh" {
			for _, item := range city.Item {
				if strings.Contains(item.Type, "SJC") { // Chỉ lấy các loại vàng SJC
					buyPrice, _ := strconv.ParseFloat(item.Buy, 64)
					sellPrice, _ := strconv.ParseFloat(item.Sell, 64)
					result.WriteString(fmt.Sprintf("**%s**\n", item.Type))
					result.WriteString(fmt.Sprintf("  - Mua: `%s`\n", formatFloat(buyPrice*1000)))
					result.WriteString(fmt.Sprintf("  - Bán: `%s`\n", formatFloat(sellPrice*1000)))
				}
			}
			break
		}
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

// Lấy tỷ giá JPY/VND (từ Vietcombank)
func getJpyVndRate() (string, error) {
	resp, err := http.Get("https://portal.vietcombank.com.vn/Usercontrols/TV_TyGia/TyGia.aspx")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	byteValue, _ := io.ReadAll(resp.Body)
    var vcbData VcbExrateList
    xml.Unmarshal(byteValue, &vcbData)

	for _, rate := range vcbData.Exrate {
		if rate.CurrencyCode == "JPY" {
			return fmt.Sprintf("🇯🇵/🇻🇳 **Tỷ giá JPY/VND (Vietcombank):**\n  - Mua (chuyển khoản): `%s VND`\n  - Bán: `%s VND`", rate.Transfer, rate.Sell), nil
		}
	}
	
	return "", fmt.Errorf("không tìm thấy tỷ giá JPY")
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
	
	// Dùng Markdown để format cho đẹp
	payload := fmt.Sprintf(`{"chat_id":%d, "text":"%s", "parse_mode":"Markdown"}`, chatID, text)

	_, err := http.Post(apiURL, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Printf("Error sending message to Telegram: %v", err)
	}
}

// Hàm handler chính mà Vercel sẽ gọi
func Handler(w http.ResponseWriter, r *http.Request) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if update.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var responseText string
	var err error

	// Phân tích lệnh từ người dùng
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
