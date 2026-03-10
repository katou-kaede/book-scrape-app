import { useState, useEffect } from 'react';
import { fetchBooks, startScraping } from './api';

// データの型定義
interface Book {
  id: number;
  title: string;
  price: string;
  stock: string;
}

function App() {
  const [books, setBooks] = useState<Book[]>([]);
  const [loading, setLoading] = useState(false);

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
    setLoading(true);
    try {
      await startScraping();
      alert("スクレイピングを開始しました！裏で動いています。");
      // 少し待ってからリストを更新する（Go側で少しデータが入るのを待つ）
      setTimeout(loadData, 2000);
    } catch (error: any) {
      if (error.response?.status === 409) {
        alert("現在実行中です。");
      }
    } finally {
      setLoading(false);
    }
  };

  // 画面が開いたときに一度データを読み込む
  useEffect(() => {
    loadData();
  }, []);

  return (
    <div style={{ padding: '20px' }}>
      <h1>📚 Book Scraper</h1>
      
      <button onClick={handleScrape} disabled={loading}>
        {loading ? "実行中..." : "最新データを取得（スクレイピング開始）"}
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
          {books.map((book) => (
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