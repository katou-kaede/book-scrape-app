package scraper

import (
	"log"
	"regexp"
	// "fmt"
	//"context"
	//"time"

	"book-scrape-app/backend/internal/model"
	"book-scrape-app/backend/internal/repository"

	"github.com/playwright-community/playwright-go"
)

type Scraper struct {
	repo *repository.BookRepository
	isScanning bool
}

func NewScraper(repo *repository.BookRepository) *Scraper {
	return &Scraper{repo: repo}
}

func (s *Scraper) Start() error {
	// すでに実行中ならスキップする
	if s.isScanning {
		log.Println("スクレイピングはすでに実行中です。")
		return nil
	}

	// 実行開始時にフラグを立てる
	s.isScanning = true
	// 関数が終わる時に必ずフラグを折る（deferを使うのが確実）
	defer func() {
		s.isScanning = false
	}()

	// 1. Playwrightの起動
	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	// 現場の鉄則: defer で確実に停止させる（リソースリーク防止）
	defer pw.Stop()

	// 2. ブラウザの起動 
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return err
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return err
	}

	log.Println("books.toscrape.comへ移動...")
	if _, err := page.Goto("https://books.toscrape.com/"); err != nil {
		return err
	}

	// 3. データの抽出
	// 在庫数（数字だけ）を抜き出すための正規表現パターン
	re := regexp.MustCompile(`\d+`)

	productPods := page.Locator(".product_pod")
	count, _ := productPods.Count()

	for i := 0; i < count; i++ {
		book := &model.Book{}
		pod := productPods.Nth(i)

		// タイトルの取得
		titleLocator := pod.Locator("h3 a")
		book.Title, _ = titleLocator.GetAttribute("title")

		// 価格の取得
		priceLocator := pod.Locator(".price_color")
		book.Price, _ = priceLocator.InnerText()

		// 詳細ページへ移動
		if err := titleLocator.Click(); err != nil {
			log.Printf("クリックに失敗: %v", err)
			continue
		}

		// 在庫数の取得
		stockLocator := page.Locator(".product_main .instock.availability")
		if err := stockLocator.WaitFor(); err != nil {
			log.Printf("詳細ページの在庫要素が見つかりません: %v", err)
		}

		stockText, _ := stockLocator.InnerText()

		// 正規表現で「19」などの数字だけを抽出
		match := re.FindString(stockText)
		book.Stock = match

		// 4. Repository経由で保存
		if err := s.repo.Save(book); err != nil {
			log.Printf("保存に失敗: %v", err)
			// 一つの保存失敗で全体を止めないのが現場流
			continue
		}

		// 5. 前のページに戻る（ブラウザの「戻る」ボタンと同じ動作）
		if _, err := page.GoBack(); err != nil {
			log.Printf("ブラウザバックに失敗: %v", err)
			break
		}
	}

	log.Printf("%d冊の本を取得しました", count)
	return nil
}


func (s *Scraper) IsProcessing() bool {
	return s.isScanning
}