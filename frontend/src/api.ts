import axios from 'axios';

// GoのサーバーのURLを設定
const API_BASE_URL = import.meta.env.VITE_API_URL || '';

const api = axios.create({
  baseURL: API_BASE_URL,
});

export interface ScrapeStatus {
    isScanning: boolean;
    lastError: string;
    currentCount: number;
    totalCount: number;
}

// 本の一覧を取得する関数
export const fetchBooks = async () => {
  const response = await api.get('/books');
  return response.data ?? [];
};

// スクレイピングを開始する関数
export const startScraping = async () => {
  const response = await api.post('/scrape');
  return response.data;
};

export const getScrapeStatus = async (): Promise<ScrapeStatus> => {
  const response = await api.get<ScrapeStatus>(`/scrape/status`);
  return response.data;
}

export const downloadCSV = async (): Promise<Blob> => {
  const response = await api.get('/books/download', {
    responseType: 'blob', // バイナリデータとして受け取る
  });
  return response.data;
}