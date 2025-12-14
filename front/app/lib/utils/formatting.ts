// フォーマット処理ユーティリティ

// 日付のフォーマット
export const formatDate = (date: Date | string, format: 'short' | 'long' | 'time' = 'short'): string => {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  
  if (isNaN(dateObj.getTime())) {
    return '無効な日付';
  }
  
  const options: Intl.DateTimeFormatOptions = {
    timeZone: 'Asia/Tokyo',
  };
  
  switch (format) {
    case 'short':
      options.year = 'numeric';
      options.month = '2-digit';
      options.day = '2-digit';
      break;
    case 'long':
      options.year = 'numeric';
      options.month = 'long';
      options.day = 'numeric';
      options.weekday = 'long';
      break;
    case 'time':
      options.hour = '2-digit';
      options.minute = '2-digit';
      options.second = '2-digit';
      break;
  }
  
  return dateObj.toLocaleDateString('ja-JP', options);
};

// 相対時間のフォーマット
export const formatRelativeTime = (date: Date | string): string => {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - dateObj.getTime();
  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  
  if (diffMinutes < 1) {
    return 'たった今';
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分前`;
  } else if (diffHours < 24) {
    return `${diffHours}時間前`;
  } else if (diffDays < 7) {
    return `${diffDays}日前`;
  } else {
    return formatDate(dateObj, 'short');
  }
};

// 数値のフォーマット
export const formatNumber = (num: number, options?: Intl.NumberFormatOptions): string => {
  return num.toLocaleString('ja-JP', options);
};

// 通貨のフォーマット
export const formatCurrency = (amount: number, currency: string = 'JPY'): string => {
  return formatNumber(amount, {
    style: 'currency',
    currency: currency,
  });
};

// パーセンテージのフォーマット
export const formatPercentage = (value: number, decimals: number = 1): string => {
  return formatNumber(value, {
    style: 'percent',
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
};

// ファイルサイズのフォーマット
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
};

// 電話番号のフォーマット
export const formatPhoneNumber = (phoneNumber: string): string => {
  const cleaned = phoneNumber.replace(/\D/g, '');
  
  if (cleaned.length === 10) {
    return cleaned.replace(/(\d{3})(\d{3})(\d{4})/, '$1-$2-$3');
  } else if (cleaned.length === 11) {
    return cleaned.replace(/(\d{3})(\d{4})(\d{4})/, '$1-$2-$3');
  }
  
  return phoneNumber;
};

// 郵便番号のフォーマット
export const formatPostalCode = (postalCode: string): string => {
  const cleaned = postalCode.replace(/\D/g, '');
  
  if (cleaned.length === 7) {
    return cleaned.replace(/(\d{3})(\d{4})/, '$1-$2');
  }
  
  return postalCode;
};

// テキストの切り詰め
export const truncateText = (text: string, maxLength: number, suffix: string = '...'): string => {
  if (text.length <= maxLength) {
    return text;
  }
  
  return text.substring(0, maxLength - suffix.length) + suffix;
};

// キャメルケースをケバブケースに変換
export const camelToKebab = (str: string): string => {
  return str.replace(/([a-z0-9]|(?=[A-Z]))([A-Z])/g, '$1-$2').toLowerCase();
};

// ケバブケースをキャメルケースに変換
export const kebabToCamel = (str: string): string => {
  return str.replace(/-([a-z])/g, (match, letter) => letter.toUpperCase());
};

// スネークケースをキャメルケースに変換
export const snakeToCamel = (str: string): string => {
  return str.replace(/_([a-z])/g, (match, letter) => letter.toUpperCase());
};

// 文字列の最初の文字を大文字に
export const capitalize = (str: string): string => {
  return str.charAt(0).toUpperCase() + str.slice(1);
};

// 文字列をタイトルケースに
export const toTitleCase = (str: string): string => {
  return str.replace(/\w\S*/g, (txt) => 
    txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase()
  );
};

// HTMLエスケープ
export const escapeHtml = (text: string): string => {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
};

// HTMLアンエスケープ
export const unescapeHtml = (html: string): string => {
  const div = document.createElement('div');
  div.innerHTML = html;
  return div.textContent || div.innerText || '';
};

// URLスラッグの生成
export const generateSlug = (text: string): string => {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '') // 特殊文字を除去
    .replace(/[\s_-]+/g, '-') // スペース、アンダースコア、ハイフンをハイフンに統一
    .replace(/^-+|-+$/g, ''); // 先頭と末尾のハイフンを除去
};

// 配列を文字列に変換（日本語対応）
export const formatList = (items: string[], conjunction: string = 'と'): string => {
  if (items.length === 0) return '';
  if (items.length === 1) return items[0];
  if (items.length === 2) return `${items[0]}${conjunction}${items[1]}`;
  
  const lastItem = items[items.length - 1];
  const otherItems = items.slice(0, -1);
  
  return `${otherItems.join('、')}${conjunction}${lastItem}`;
};

// 時間の長さをフォーマット
export const formatDuration = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = seconds % 60;
  
  if (hours > 0) {
    return `${hours}時間${minutes}分${remainingSeconds}秒`;
  } else if (minutes > 0) {
    return `${minutes}分${remainingSeconds}秒`;
  } else {
    return `${remainingSeconds}秒`;
  }
};
