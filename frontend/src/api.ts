import axios from 'axios';

// GoのサーバーのURLを設定
const API_BASE_URL = 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// 本の一覧を取得する関数
export const fetchBooks = async () => {
  const response = await api.get('/books');
  return response.data;
};

// スクレイピングを開始する関数
export const startScraping = async () => {
  const response = await api.post('/scrape');
  return response.data;
};