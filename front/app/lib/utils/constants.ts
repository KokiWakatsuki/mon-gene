// アプリケーション定数

// API関連
export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/auth/login',
    LOGOUT: '/auth/logout',
    REGISTER: '/auth/register',
    VERIFY: '/auth/verify',
    REFRESH: '/auth/refresh',
    FORGOT_PASSWORD: '/auth/forgot-password',
    RESET_PASSWORD: '/auth/reset-password',
  },
  PROBLEMS: {
    LIST: '/problems',
    GENERATE: '/problems/generate',
    DETAIL: (id: string) => `/problems/${id}`,
    DELETE: (id: string) => `/problems/${id}`,
    PDF: (id: string) => `/problems/${id}/pdf`,
    FAVORITES: '/problems/favorites',
  },
  HEALTH: '/health',
} as const;

// ローカルストレージキー
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'authToken',
  USER_DATA: 'userData',
  THEME: 'theme',
  LANGUAGE: 'language',
  FILTERS: 'selectedFilters',
  PREFERENCES: 'userPreferences',
} as const;

// 科目定数
export const SUBJECTS = {
  MATH: '数学',
  ENGLISH: '英語',
  JAPANESE: '国語',
  SCIENCE: '理科',
  SOCIAL: '社会',
} as const;

// 科目別単元
export const SUBJECT_UNITS = {
  [SUBJECTS.MATH]: [
    { label: '式の計算', value: 'calculation' },
    { label: '図形', value: 'geometry' },
    { label: '空間図形', value: 'spatial_geometry' },
    { label: '2次不等式', value: 'quadratic' },
    { label: '関数', value: 'function' },
    { label: '確率', value: 'probability' },
  ],
  [SUBJECTS.ENGLISH]: [
    { label: '文法', value: 'grammar' },
    { label: '読解', value: 'reading' },
    { label: '語彙', value: 'vocabulary' },
    { label: 'リスニング', value: 'listening' },
  ],
  [SUBJECTS.JAPANESE]: [
    { label: '現代文', value: 'modern' },
    { label: '古文', value: 'classical' },
    { label: '漢文', value: 'chinese' },
    { label: '文法', value: 'grammar' },
  ],
} as const;

// 学年
export const GRADES = [
  { label: '中1', value: 'grade1' },
  { label: '中2', value: 'grade2' },
  { label: '中3', value: 'grade3' },
  { label: '高1', value: 'grade4' },
  { label: '高2', value: 'grade5' },
  { label: '高3', value: 'grade6' },
] as const;

// 難易度
export const DIFFICULTY_LEVELS = [
  { label: 'Lv1', value: 'level1' },
  { label: 'Lv2', value: 'level2' },
  { label: 'Lv3', value: 'level3' },
  { label: 'Lv4', value: 'level4' },
  { label: 'Lv5', value: 'level5' },
] as const;

// 必要な公式数
export const FORMULA_COUNTS = [
  { label: '1個', value: 'formula1' },
  { label: '2個', value: 'formula2' },
  { label: '3個', value: 'formula3' },
  { label: '4個以上', value: 'formula4plus' },
] as const;

// 計算量
export const CALCULATION_COMPLEXITY = [
  { label: '簡単', value: 'simple' },
  { label: '普通', value: 'medium' },
  { label: '複雑', value: 'complex' },
] as const;

// 数値の複雑性
export const NUMBER_COMPLEXITY = [
  { label: '整数のみ', value: 'integer' },
  { label: '小数を含む', value: 'decimal' },
  { label: '分数を含む', value: 'fraction' },
] as const;

// 問題文の文章量
export const TEXT_LENGTH = [
  { label: '短い', value: 'short' },
  { label: '普通', value: 'medium' },
  { label: '長い', value: 'long' },
] as const;

// テーマ設定
export const THEMES = {
  LIGHT: 'light',
  DARK: 'dark',
  SYSTEM: 'system',
} as const;

// 言語設定
export const LANGUAGES = {
  JAPANESE: 'ja',
  ENGLISH: 'en',
} as const;

// ページサイズ
export const PAGE_SIZES = {
  SMALL: 10,
  MEDIUM: 20,
  LARGE: 50,
} as const;

// ファイルサイズ制限
export const FILE_SIZE_LIMITS = {
  IMAGE: 5 * 1024 * 1024, // 5MB
  PDF: 10 * 1024 * 1024,  // 10MB
  DOCUMENT: 2 * 1024 * 1024, // 2MB
} as const;

// 許可されるファイル形式
export const ALLOWED_FILE_TYPES = {
  IMAGE: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  PDF: ['application/pdf'],
  DOCUMENT: ['application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'],
} as const;

// HTTPステータスコード
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  UNPROCESSABLE_ENTITY: 422,
  INTERNAL_SERVER_ERROR: 500,
  SERVICE_UNAVAILABLE: 503,
} as const;

// エラーメッセージ
export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'ネットワークエラーが発生しました',
  UNAUTHORIZED: '認証が必要です',
  FORBIDDEN: 'アクセス権限がありません',
  NOT_FOUND: 'リソースが見つかりません',
  VALIDATION_ERROR: '入力内容に誤りがあります',
  SERVER_ERROR: 'サーバーエラーが発生しました',
  UNKNOWN_ERROR: '不明なエラーが発生しました',
} as const;

// 成功メッセージ
export const SUCCESS_MESSAGES = {
  LOGIN_SUCCESS: 'ログインしました',
  LOGOUT_SUCCESS: 'ログアウトしました',
  SAVE_SUCCESS: '保存しました',
  DELETE_SUCCESS: '削除しました',
  UPDATE_SUCCESS: '更新しました',
  GENERATE_SUCCESS: '問題を生成しました',
} as const;

// アニメーション設定
export const ANIMATION_DURATION = {
  FAST: 150,
  NORMAL: 300,
  SLOW: 500,
} as const;

// ブレークポイント
export const BREAKPOINTS = {
  SM: 640,
  MD: 768,
  LG: 1024,
  XL: 1280,
  '2XL': 1536,
} as const;

// Z-index値
export const Z_INDEX = {
  DROPDOWN: 1000,
  STICKY: 1020,
  FIXED: 1030,
  MODAL_BACKDROP: 1040,
  MODAL: 1050,
  POPOVER: 1060,
  TOOLTIP: 1070,
  TOAST: 1080,
} as const;

// デバウンス時間（ミリ秒）
export const DEBOUNCE_DELAY = {
  SEARCH: 300,
  RESIZE: 100,
  SCROLL: 50,
} as const;

// 正規表現パターン
export const REGEX_PATTERNS = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  PHONE: /^[\d-+().\s]+$/,
  POSTAL_CODE: /^\d{3}-?\d{4}$/,
  SCHOOL_CODE: /^[A-Za-z0-9]+$/,
  PASSWORD: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$/,
} as const;

// デフォルト設定
export const DEFAULT_SETTINGS = {
  THEME: THEMES.SYSTEM,
  LANGUAGE: LANGUAGES.JAPANESE,
  PAGE_SIZE: PAGE_SIZES.MEDIUM,
  AUTO_SAVE: true,
  NOTIFICATIONS: true,
} as const;
