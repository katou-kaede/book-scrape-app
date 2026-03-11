import axios from 'axios';

// GoのサーバーのURLを設定
const API_BASE_URL = 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
});

export interface ScrapeStatus {
    isScanning: boolean;
    lastError: string;
}

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

export const getScrapeStatus = async (): Promise<ScrapeStatus> => {
  const res = await axios.get<ScrapeStatus>(`${API_BASE_URL}/scrape/status`);
  return res.data;
}