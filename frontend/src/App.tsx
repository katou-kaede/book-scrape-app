import { useState, useEffect } from 'react';
import { fetchBooks, startScraping, getScrapeStatus } from './api';

// データの型定義
interface Book {
  id: number;
  title: string;
  price: string;
  stock: string;
}

function App() {
  const [books, setBooks] = useState<Book[]>([]);
  const [isScanning, setIsScanning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 1. データを読み込む関数
  const loadData = async () => {
    try {
      const data = await fetchBooks();
      setBooks(data);
    } catch (error) {
      console.error("データ取得失敗:", error);
    }
  };

  // 2. スクレイピング開始ボタンの処理
  const handleScrape = async () => {
    try {
      setError(null);
      await startScraping();
      setIsScanning(true);
      // 少し待ってからリストを更新する（Go側で少しデータが入るのを待つ）
      // setTimeout(loadData, 2000);
    } catch (error: any) {
      if (error.response?.status === 409) {
        alert("現在実行中です。");
        setIsScanning(true);
      } else {
        setError("サーバーとの通信に失敗しました。");
      }
    }
  };

  // ポーリング処理
  useEffect(() => {
    let intervalId: number;

    if (isScanning) {
      intervalId = window.setInterval(async () => {
        try {
          const status = await getScrapeStatus();
          
          if (!status.isScanning) {
            // サーバー側で終わった場合
            setIsScanning(false);
            clearInterval(intervalId);
            loadData(); // 終わったので最新化
            
            if (status.lastError) {
              setError(status.lastError);
            }
          }
        } catch (err) {
          console.error("ステータス確認失敗", err);
        }
      }, 2000);
    }

    return () => clearInterval(intervalId);
  }, [isScanning]); // isScanningが変わるたびに監視を開始/停止

  // 画面が開いたときに一度データを読み込む
  useEffect(() => {
    loadData();
  }, []);

  return (
    <div style={{ padding: '20px' }}>
      <h1>📚 Book Scraper</h1>

      {/* エラーメッセージの表示 */}
      {error && (
        <div style={{ color: 'red', backgroundColor: '#ffdada', padding: '10px', marginBottom: '10px', borderRadius: '5px' }}>
          <strong>エラー:</strong> {error}
        </div>
      )}
      
      <button onClick={handleScrape} disabled={isScanning}>
        {isScanning ? "実行中..." : "最新データを取得（スクレイピング開始）"}
      </button>

      <button onClick={loadData} style={{ marginLeft: '10px' }}>
        表示を更新
      </button>

      <hr />

      <table border={1} style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ backgroundColor: '#f4f4f4' }}>
            <th>ID</th>
            <th>タイトル</th>
            <th>価格</th>
            <th>在庫状況</th>
          </tr>
        </thead>
        <tbody>
          {books && books.map((book) => (
            <tr key={book.id}>
              <td>{book.id}</td>
              <td>{book.title}</td>
              <td>{book.price}</td>
              <td>{book.stock}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default App;