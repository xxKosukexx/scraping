package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"os"

	"log"
	"net/http"
	"net/smtp"
	"strings"
)

// ギークスの案件最新10件を取得
func getGeechsLatestInfos() []string {
	res, err := http.Get("https://geechs-job.com/project")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// goqueryでドキュメントを解析
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// ドキュメント内の特定の要素を抽出
	latestInfos := []string{}
	doc.Find(".c-card_title_link").Each(func(i int, s *goquery.Selection) {
		latestInfos = append(latestInfos, s.Text())
	})

	return latestInfos
}

// Gmailでギークスの案件情報をメール送信
func sendGmail(latestInfos []string) {
	// SMTPサーバーの設定
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// 認証情報
	sender := os.Getenv("GMAIL_ADDRESS")
	password := os.Getenv("GMAIL_APP_PASSWORD") // Gmailのアプリパスワード

	// メールの内容
	receiver := os.Getenv("GMAIL_ADDRESS")
	subject := "Geechs Jobの最新案件情報です\n"
	body := fmt.Sprintf("geechs jobの最新案件10件です。\n%s", strings.Join(latestInfos, "\n"))

	message := []byte(subject + "\n" + body)

	// TLSによるSMTP接続
	auth := smtp.PlainAuth("", sender, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{receiver}, message)
	if err != nil {
		panic(err)
	}
}

// .envファイルを読み込む
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	loadEnv()
	latestInfos := getGeechsLatestInfos()
	sendGmail(latestInfos)
}