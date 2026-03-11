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
    <div className="min-h-screen bg-slate-50 p-4 md:p-8 text-slate-800">
      <div className="max-w-5xl mx-auto">
        {/* ヘッダーセクション */}
        <header className="flex justify-between items-center mb-8 bg-white p-6 rounded-2xl shadow-sm border border-slate-200">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2 text-slate-900">
              <span className="text-3xl">📚</span> Book Scraper
            </h1>
            <p className="text-sm text-slate-500 mt-1">実演用スクレイピング・ダッシュボード</p>
          </div>
          
          <div className="flex gap-3">
            <button 
              onClick={loadData} 
              className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-full transition-all cursor-pointer"
              title="表示を更新"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
            
            <button 
              onClick={handleScrape} 
              disabled={isScanning}
              className={`flex items-center gap-2 px-6 py-2.5 rounded-xl font-bold transition-all shadow-lg active:scale-95 ${
                isScanning 
                  ? "bg-slate-200 text-slate-400 cursor-not-allowed" 
                  : "bg-indigo-600 text-white hover:bg-indigo-700 shadow-indigo-200 cursor-pointer"
              }`}
            >
              {isScanning ? (
                <>
                  <div className="w-5 h-5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin"></div>
                  実行中...
                </>
              ) : "スクレイピング開始"}
            </button>
          </div>
        </header>

        {/* エラーアラート */}
        {error && (
          <div className="mb-6 flex items-center gap-3 bg-red-50 border border-red-100 text-red-700 px-4 py-3 rounded-xl animate-bounce">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">{error}</span>
          </div>
        )}

        {/* メインテーブル */}
        <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200">
                <th className="px-6 py-4 text-xs font-bold text-slate-500 uppercase tracking-wider">ID</th>
                <th className="px-6 py-4 text-xs font-bold text-slate-500 uppercase tracking-wider">タイトル</th>
                <th className="px-6 py-4 text-xs font-bold text-slate-500 uppercase tracking-wider text-right">価格</th>
                <th className="px-6 py-4 text-xs font-bold text-slate-500 uppercase tracking-wider text-center">在庫状況</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {books.length > 0 ? (
                books.map((book) => (
                  <tr key={book.id} className="hover:bg-slate-50/80 transition-colors group">
                    <td className="px-6 py-4 text-sm text-slate-400 font-mono">#{book.id}</td>
                    <td className="px-6 py-4">
                      <div className="text-sm font-semibold text-slate-800 group-hover:text-indigo-600 transition-colors">
                        {book.title}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm font-bold text-slate-900 text-right font-mono">
                      {book.price}
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        parseInt(book.stock) > 5 
                          ? "bg-emerald-100 text-emerald-700" 
                          : "bg-amber-100 text-amber-700"
                      }`}>
                        残り {book.stock} 冊
                      </span>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={4} className="px-6 py-20 text-center text-slate-400">
                    <p className="text-lg mb-1">データがありません</p>
                    <p className="text-sm">上のボタンを押してスクレイピングを開始してください</p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default App;