package handler

import (
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

// Lấy giá vàng SJC từ trang giavang.org
func getVnGoldPrice() (string, error) {
	url := "https://giavang.org/"

	res, err := makeRequest(url)
	if err != nil {
		return "", fmt.Errorf("không thể truy cập giavang.org: %v", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("lỗi đọc dữ liệu trang: %v", err)
	}

	var buyPrice, sellPrice, updateTime string

	// Tìm đến bảng giá SJC và duyệt qua từng hàng
	doc.Find("#giasjc tbody tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Lấy tên loại vàng ở cột đầu tiên
		label := s.Find("td").First().Text()

		// Chúng ta chỉ quan tâm đến loại vàng miếng phổ biến nhất
		if strings.Contains(label, "SJC 1L, 10L") {
			// Lấy giá mua ở cột thứ 2
			buyPrice = s.Find("td").Eq(1).Text()
			// Lấy giá bán ở cột thứ 3
			sellPrice = s.Find("td").Eq(2).Text()

			// Lấy thời gian cập nhật ở hàng trên cùng của bảng
			updateTime = doc.Find("#giasjc .updated").Text()

			// Đã tìm thấy, không cần duyệt nữa
			return false 
		}
		// Nếu không tìm thấy, tiếp tục duyệt
		return true
	})

	if buyPrice == "" || sellPrice == "" {
		return "", fmt.Errorf("không tìm thấy giá vàng SJC 1L trên trang (cấu trúc có thể đã thay đổi)")
	}
	
	// Format lại chuỗi kết quả cho đẹp
	var result strings.Builder
	result.WriteString("🇻🇳 **Giá Vàng SJC 1L, 10L**\n")
	result.WriteString(fmt.Sprintf("*(Nguồn: giavang.org, %s)*\n", strings.TrimSpace(updateTime)))
	result.WriteString("------------------------------------\n")
	result.WriteString(fmt.Sprintf("  - **Mua vào:** `%s.000 VND`\n", buyPrice))
	result.WriteString(fmt.Sprintf("  - **Bán ra:**   `%s.000 VND`", sellPrice))

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
