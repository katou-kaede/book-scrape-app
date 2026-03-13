package scraper

import (
	"log"
	"regexp"
	"fmt"
	"strconv"
	//"context"
	//"time"

	"book-scrape-app/backend/internal/model"
	"book-scrape-app/backend/internal/repository"

	"github.com/playwright-community/playwright-go"
)

type Scraper struct {
	repo *repository.BookRepository
	isScanning bool
	lastError string
	currentCount int
	totalCount int
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

	s.isScanning = true  // 実行開始時にフラグを立てる
	s.lastError = ""  // 前回のエラーをリセット
	s.currentCount = 0
	s.totalCount = 0

	var finalErr error
	// 関数が終わる時に必ずフラグを折る（deferを使うのが確実）
	defer func() {
		if finalErr != nil {
			s.lastError = finalErr.Error()
		}
		s.isScanning = false
	}()

	// 1. Playwrightの起動
	pw, err := playwright.Run()
	if err != nil {
		finalErr = fmt.Errorf("Playwrightの起動に失敗: %w", err)
		return finalErr
	}
	// 現場の鉄則: defer で確実に停止させる（リソースリーク防止）
	defer pw.Stop()

	// 2. ブラウザの起動 
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args: []string{
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage", // これが超重要！メモリ不足を回避します
			"--disable-gpu",           // LinuxコンテナではGPU不要
		},
	})
	if err != nil {
		finalErr = fmt.Errorf("ブラウザの起動に失敗: %w", err)
		return finalErr
	}
	defer browser.Close()

	// ブラウザを起動した後
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		// 証明書エラーを無視する設定
		IgnoreHttpsErrors: playwright.Bool(true),
		// 画面サイズをフルHDくらいに広げておく
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		log.Fatalf("could not create context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		finalErr = fmt.Errorf("新しいページの作成に失敗: %w", err)
		return finalErr
	}

	log.Println("books.toscrape.comへ移動...")
	if _, err := page.Goto("https://books.toscrape.com/"); err != nil {
		finalErr = fmt.Errorf("ページへの移動に失敗: %w", err)
		return finalErr
	}

	// 3. データの抽出
	// 在庫数（数字だけ）を抜き出すための正規表現パターン
	re := regexp.MustCompile(`\d+`)

	// ページ全体の読み込みを待つ
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	err = page.Locator(".product_pod").First().WaitFor()
	if err != nil {
		log.Printf("要素の待機中にエラーが発生しました: %v", err)
	}

	maxPages := s.GetMaxPages(page)
	// maxPages := 3
	productPods := page.Locator(".product_pod")
	count, _ := productPods.Count()
	s.totalCount = count * maxPages // 全体の冊数を概算

	for p := 1; p <= maxPages; p++ {
		targetURL := fmt.Sprintf("http://books.toscrape.com/catalogue/page-%d.html", p)
		if _, err := page.Goto(targetURL); err != nil {
            continue
        }

		// ページ全体の読み込みを待つ
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded,
		})

		err = page.Locator(".product_pod").First().WaitFor()
		if err != nil {
			log.Printf("要素の待機中にエラーが発生しました: %v", err)
		}

		productPods := page.Locator(".product_pod")
		count, _ := productPods.Count()
	
		for i := 0; i < count; i++ {
			// ループの冒頭で「今何冊目か」を更新
			s.currentCount++

			book := &model.Book{}
			pod := productPods.Nth(i)

			// ページ全体の読み込みを待つ
			page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			})

			err = page.Locator(".product_pod").First().WaitFor()
			if err != nil {
				log.Printf("要素の待機中にエラーが発生しました: %v", err)
			}

			// タイトルの取得
			titleLocator := pod.Locator("h3 a")
			book.Title, _ = titleLocator.GetAttribute("title")

			// 価格の取得
			priceLocator := pod.Locator(".price_color")
			book.Price, _ = priceLocator.TextContent()

			// 詳細ページへ移動
			//  まず、要素が見える位置までスクロールさせる
			// if err := titleLocator.ScrollIntoViewIfNeeded(); err != nil {
			// 	log.Printf("スクロール失敗: %v", err)
			// }
			// if err := titleLocator.Click(playwright.LocatorClickOptions{
			// 	Force:   playwright.Bool(true),
    		// 	Timeout: playwright.Float(15000),
			// }); err != nil {
			// 	log.Printf("クリックに失敗: %v", err)
			// }

			if err := titleLocator.DispatchEvent("click", nil); err != nil {
				log.Printf("[%d冊目] イベント発火に失敗: %v", s.currentCount, err)
				continue
			}

			// 移動を確実に待つ（これがないと次でコケる）
			err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateNetworkidle,
			})
			if err != nil {
				log.Printf("読み込み待ちでエラー: %v", err)
			}

			// 在庫数の取得
			stockLocator := page.Locator(".product_main .instock.availability")
			if err := stockLocator.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}); err != nil {
				log.Printf("詳細ページの在庫要素が見つかりません: %v", err)
			}

			stockText, _ := stockLocator.TextContent()

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
				finalErr = fmt.Errorf("ブラウザバックに失敗: %w", err)
				break
			}
		}
	}

	log.Printf("%d冊の本を取得しました", s.currentCount)
	return finalErr
}


func (s *Scraper) GetMaxPages(page playwright.Page) int {
    // セレクターで要素を取得
    text, err := page.Locator(".pager .current").TextContent()
    if err != nil {
        return 1 // 失敗したら安全のために1ページだけにする
    }

    // "Page 1 of 50" から数字をすべて抽出
    re := regexp.MustCompile(`\d+`)
    matches := re.FindAllString(text, -1)

    // matches[0] が "1"、matches[1] が "50" になるはず
    if len(matches) >= 2 {
        max, _ := strconv.Atoi(matches[1])
        return max
    }

    return 1
}


func (s *Scraper) GetStatus() (bool, string, int, int) {
    return s.isScanning, s.lastError, s.currentCount, s.totalCount
}