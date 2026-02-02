package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

	// Dùng Markdown để format cho đẹp
	payload := fmt.Sprintf(`{"chat_id":%d, "text":"%s", "parse_mode":"Markdown"}`, chatID, text)

	_, err := http.Post(apiURL, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Printf("Error sending message to Telegram: %v", err)
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

		price, err := getVnGoldPrice()
		if err != nil {
			log.Printf("Error getting gold price for cron: %v", err)
			// Vẫn báo OK để Vercel không retry liên tục nếu lỗi do nguồn
			w.WriteHeader(http.StatusOK)
			return
		}

		sendTelegramMessage(chatID, price)
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
