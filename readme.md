# Bot Telegram Tra Cứu Giá Bằng Go Trên Vercel

Đây là một dự án bot Telegram đơn giản được viết bằng ngôn ngữ Go, triển khai dưới dạng Serverless Function trên Vercel. Bot cho phép người dùng tra cứu nhanh giá Bitcoin, giá vàng và tỷ giá ngoại tệ.

## ✨ Tính Năng

Bot hỗ trợ các lệnh sau:

-   `/start`: Hiển thị tin nhắn chào mừng.
-   `/bitcoin`: Trả về giá Bitcoin (USD) mới nhất từ CoinGecko.
-   `/vang`: Trả về giá vàng thế giới (USD/oz) từ GoldPrice.org.
-   `/vangvn`: Trả về giá vàng SJC (VND) tại TP.HCM.
-   `/usdjpy`: Trả về tỷ giá USD/JPY.
-   `/jpyvnd`: Trả về tỷ giá JPY/VND từ Vietcombank.

## 🛠️ Công Nghệ Sử Dụng

-   **Ngôn ngữ:** [Go](https://go.dev/)
-   **Nền tảng triển khai:** [Vercel](https://vercel.com/) (Serverless Functions)
-   **Nền tảng Bot:** [Telegram Bot API](https://core.telegram.org/bots/api)
-   **Thư viện Go:**
    -   `net/http` (chuẩn của Go)
    -   `encoding/json`, `encoding/xml` (chuẩn của Go)
    -   `github.com/PuerkitoBio/goquery` (để cào dữ liệu web)

---

## 🚀 Hướng Dẫn Cài Đặt và Triển Khai

Thực hiện theo các bước sau để triển khai bot của riêng bạn.

### Bước 1: Điều Kiện Tiên Quyết

-   Cài đặt [Go](https://go.dev/doc/install) (phiên bản 1.18 trở lên).
-   Cài đặt [Node.js và npm](https://nodejs.org/en/) (để cài Vercel CLI).
-   Cài đặt [Git](https://git-scm.com/).
-   Một tài khoản [Telegram](https://telegram.org/).
-   Một tài khoản [Vercel](https://vercel.com/signup) (liên kết với Github/Gitlab).

### Bước 2: Tạo Bot trên Telegram

1.  Mở Telegram, tìm kiếm `BotFather` (bot có dấu tick xanh).
2.  Gõ `/newbot` và làm theo hướng dẫn để đặt tên và username cho bot.
3.  **LƯU LẠI** token API mà BotFather cung cấp. Đây là thông tin cực kỳ quan trọng và cần được giữ bí mật.

### Bước 3: Chuẩn Bị Mã Nguồn

1.  Tạo một thư mục cho dự án.
2.  Tạo cấu trúc thư mục như sau:

    ```
    /your-bot-project
    |-- /api
    |   |-- handler.go
    |-- go.mod
    |-- go.sum
    |-- vercel.json
    ```

### Bước 4: Viết Mã Nguồn và Cấu Hình

**1. File `api/handler.go`:**
Dán toàn bộ code Go đã được cung cấp vào file này.

**2. Khởi tạo Go Modules:**
Mở terminal trong thư mục gốc của dự án và chạy các lệnh:
```bash
# Khởi tạo module
go mod init ten-du-an-cua-ban

# Tải thư viện goquery
go get github.com/PuerkitoBio/goquery
```
Lệnh này sẽ tự động tạo ra hai file `go.mod` và `go.sum`.

**3. File `vercel.json`:**
Tạo file `vercel.json` với nội dung sau. File này báo cho Vercel biết cách build và chạy code Go của bạn.
```json
{
  "builds": [
    {
      "src": "api/handler.go",
      "use": "@vercel/go"
    }
  ],
  "rewrites": [
    {
      "source": "/api/handler",
      "destination": "/api/handler.go"
    }
  ]
}
```

### Bước 5: Triển Khai Lên Vercel

1.  **Đưa code lên Github:**
    -   Khởi tạo Git: `git init`
    -   Tạo một kho chứa mới trên Github.
    -   Thêm, commit và đẩy code của bạn lên kho chứa đó.

2.  **Import Dự Án vào Vercel:**
    -   Truy cập [Vercel Dashboard](https://vercel.com/dashboard).
    -   Chọn "Add New..." -> "Project".
    -   Chọn kho chứa Github bạn vừa tạo. Vercel sẽ tự nhận diện đây là dự án Go.

3.  **Thiết Lập Biến Môi Trường (Rất Quan Trọng):**
    -   Trong quá trình import, tìm đến mục **Environment Variables**.
    -   Thêm một biến mới:
        -   **Name:** `TELEGRAM_TOKEN`
        -   **Value:** Dán token bot của bạn vào đây.
    -   Nhấn **Deploy**. Vercel sẽ bắt đầu quá trình build và triển khai.

### Bước 6: Kết Nối Bot với Vercel (Set Webhook)

Sau khi Vercel triển khai xong, bạn sẽ có một URL (ví dụ: `https://your-bot.vercel.app`).

1.  Lấy URL đó và ghép với đường dẫn đã cấu hình: `https://your-bot.vercel.app/api/handler`.
2.  Lấy token bot của bạn.
3.  Mở trình duyệt và truy cập vào URL sau (thay thế các giá trị trong `< >`):

    ```
    https://api.telegram.org/bot<TOKEN_CUA_BAN>/setWebhook?url=<URL_VERCEL_CUA_BAN>/api/handler
    ```
4.  Nếu trình duyệt trả về: `{"ok":true,"result":true,"description":"Webhook was set"}`, bạn đã thành công!

### Bước 7: Thiết Lập Tự Động Cập Nhật (Cron Job)

Vì Vercel Free Tier giới hạn cron job, chúng ta dùng **GitHub Actions** để gọi bot mỗi giờ. Để bảo mật URL, chúng ta dùng GitHub Secrets.

1.  **Lấy URL Cron của bạn:**
    URL sẽ có dạng: `https://<TEN_DU_AN_CUA_BAN>.vercel.app/api/handler?mode=cron`

2.  **Thêm Secret trên Github:**
    -   Vào repository của bạn trên Github.
    -   Chọn **Settings** -> **Secrets and variables** -> **Actions**.
    -   Nhấn **New repository secret**.
    -   **Name**: `CRON_URL`
    -   **Secret**: Dán URL Cron của bạn vào đây.
    -   Nhấn **Add secret**.

3.  **Kích hoạt:**
    GitHub Actions sẽ tự động chạy theo lịch trình. Bạn có thể kiểm tra trong tab **Actions** của repository.

---

## 🐞 Chẩn Đoán và Sửa Lỗi

### Vấn Đề 1: Bot không phản hồi bất cứ lệnh nào.

-   **Triệu chứng:** Bạn gửi lệnh `/start` nhưng bot "im re".
-   **Chẩn đoán:** Rất có thể Webhook chưa được cài đặt hoặc cài đặt sai.
-   **Hành động:**
    1.  Mở trình duyệt và truy cập URL sau để kiểm tra: `https://api.telegram.org/bot<TOKEN_CUA_BAN>/getWebhookInfo`
    2.  Xem kết quả. Nếu trường `"url"` rỗng (`"url": ""`), nghĩa là Webhook chưa được cài.
    3.  **Giải pháp:** Thực hiện lại **Bước 6** một cách cẩn thận. Đảm bảo URL Vercel và token không bị gõ nhầm.

### Vấn Đề 2: Bot không phản hồi và `getWebhookInfo` báo lỗi "500 Internal Server Error".

-   **Triệu chứng:** `getWebhookInfo` trả về một lỗi trong trường `"last_error_message"`:
    ```json
    "last_error_message": "Wrong response from the webhook: 500 Internal Server Error"
    ```
-   **Chẩn đoán:** Webhook đã được cài đúng! Telegram đã gửi yêu cầu thành công, nhưng **code Go của bạn đã bị crash** trên máy chủ Vercel. Vấn đề nằm ở code hoặc cấu hình Vercel.

-   **Hành động:** Xem "hộp đen" của ứng dụng - Vercel Logs.
    1.  Truy cập dự án của bạn trên Vercel.
    2.  Vào tab **Logs**.
    3.  Chọn tab con **Functions**.
    4.  Gửi một lệnh cho bot trên Telegram.
    5.  **Quan sát ngay lập tức** cửa sổ log trên Vercel. Bạn sẽ thấy một thông báo lỗi màu đỏ.
    6.  **Giải pháp phổ biến nhất:**
        -   **Lỗi `Fatal error: TELEGRAM_TOKEN environment variable not set`:** Bạn đã quên thiết lập Biến Môi Trường ở **Bước 5**. Hãy vào `Settings -> Environment Variables` trên Vercel, thêm biến `TELEGRAM_TOKEN` và triển khai lại.
        -   **Các lỗi khác (`panic`, `index out of range`,...):** Đọc kỹ thông báo lỗi trong log. Nó sẽ chỉ ra chính xác dòng code nào trong `handler.go` đang gây ra vấn đề để bạn sửa lại.

