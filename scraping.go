package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"os"
	"sync"

	"log"
	"net/http"
	"net/smtp"
	"strings"
)

// 日経新聞の最新10件を取得
func getNikkeiLatestInfos(url string, resultChan chan []string, wg *sync.WaitGroup) {
	defer wg.Done()
	res, err := http.Get(url)
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
	doc.Find("[class*=fauxBlockLink]").Each(func(i int, s *goquery.Selection) {
		latestInfos = append(latestInfos, s.Text())
	})

	resultChan <- latestInfos
	return
}

// Gmailで日経新聞の情報をメール送信
func sendGmail(latestInfos []string) {
	// SMTPサーバーの設定
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// 認証情報
	sender := os.Getenv("GMAIL_ADDRESS")
	password := os.Getenv("GMAIL_APP_PASSWORD") // Gmailのアプリパスワード

	// メールの内容
	receiver := os.Getenv("GMAIL_ADDRESS")
	subject := "日本経済新聞の最新情報です\n"
	body := fmt.Sprintf("日本経済新聞の最新%d件です。\n\n%s", len(latestInfos), strings.Join(latestInfos, "\n\n"))

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
	var wg sync.WaitGroup
	resultChan := make(chan []string, 3)
	wg.Add(3)
	go getNikkeiLatestInfos("https://www.nikkei.com/economy/economy/", resultChan, &wg)
	go getNikkeiLatestInfos("https://www.nikkei.com/economy/economy/?page=2", resultChan, &wg)
	go getNikkeiLatestInfos("https://www.nikkei.com/economy/economy/?page=3", resultChan, &wg)

	wg.Wait()
	close(resultChan)

	latestInfos := []string{}
	for result := range resultChan {
		latestInfos = append(latestInfos, result...)
	}

	sendGmail(latestInfos)
}
