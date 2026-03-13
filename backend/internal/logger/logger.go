package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func SetupLogFile() func() {
	logDir := os.Getenv("LOG_DIR")
    if logDir == "" {
        logDir = "/app/logs"
    }

	// フォルダがなければ作成
    if _, err := os.Stat(logDir); os.IsNotExist(err) {
        _ = os.MkdirAll(logDir, 0755)
    }

	logFileName := filepath.Join(logDir, "app.log")
    oldLogFileName := filepath.Join(logDir, "app.old.log")

	// 1. ローテーション：古いログがあればバックアップに回す
	if _, err := os.Stat(logFileName); err == nil {
		_ = os.Remove(oldLogFileName) // 古いバックアップを削除
		_ = os.Rename(logFileName, oldLogFileName) // 現在のログをバックアップへ
	}

	// 2. 新規ログファイルの作成
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return func() {}
	}

	// 3. 出力先を「コンソール」と「ファイル」の両方に設定
	multi := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multi)

	return func() {
		f.Close()
	}
}