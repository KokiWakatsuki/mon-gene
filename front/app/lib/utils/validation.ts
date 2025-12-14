// バリデーションユーティリティ

export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

// メールアドレスのバリデーション
export const validateEmail = (email: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!email) {
    errors.push('メールアドレスは必須です');
  } else {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      errors.push('有効なメールアドレスを入力してください');
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// パスワードのバリデーション
export const validatePassword = (password: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!password) {
    errors.push('パスワードは必須です');
  } else {
    if (password.length < 8) {
      errors.push('パスワードは8文字以上である必要があります');
    }
    if (!/(?=.*[a-z])/.test(password)) {
      errors.push('パスワードには小文字を含める必要があります');
    }
    if (!/(?=.*[A-Z])/.test(password)) {
      errors.push('パスワードには大文字を含める必要があります');
    }
    if (!/(?=.*\d)/.test(password)) {
      errors.push('パスワードには数字を含める必要があります');
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 塾コードのバリデーション
export const validateSchoolCode = (schoolCode: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!schoolCode) {
    errors.push('塾コードは必須です');
  } else {
    if (schoolCode.length < 3) {
      errors.push('塾コードは3文字以上である必要があります');
    }
    if (!/^[A-Za-z0-9]+$/.test(schoolCode)) {
      errors.push('塾コードは英数字のみ使用できます');
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 必須フィールドのバリデーション
export const validateRequired = (value: string, fieldName: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!value || value.trim() === '') {
    errors.push(`${fieldName}は必須です`);
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 文字数制限のバリデーション
export const validateLength = (
  value: string, 
  min: number, 
  max: number, 
  fieldName: string
): ValidationResult => {
  const errors: string[] = [];
  
  if (value.length < min) {
    errors.push(`${fieldName}は${min}文字以上である必要があります`);
  }
  if (value.length > max) {
    errors.push(`${fieldName}は${max}文字以下である必要があります`);
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 数値のバリデーション
export const validateNumber = (
  value: string, 
  min?: number, 
  max?: number, 
  fieldName: string = '値'
): ValidationResult => {
  const errors: string[] = [];
  
  const numValue = parseFloat(value);
  
  if (isNaN(numValue)) {
    errors.push(`${fieldName}は有効な数値である必要があります`);
  } else {
    if (min !== undefined && numValue < min) {
      errors.push(`${fieldName}は${min}以上である必要があります`);
    }
    if (max !== undefined && numValue > max) {
      errors.push(`${fieldName}は${max}以下である必要があります`);
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 複数のバリデーション結果をマージ
export const mergeValidationResults = (...results: ValidationResult[]): ValidationResult => {
  const allErrors = results.flatMap(result => result.errors);
  
  return {
    isValid: allErrors.length === 0,
    errors: allErrors
  };
};

// フォーム全体のバリデーション
export const validateForm = (
  data: Record<string, any>, 
  rules: Record<string, (value: any) => ValidationResult>
): ValidationResult => {
  const results = Object.entries(rules).map(([field, validator]) => 
    validator(data[field])
  );
  
  return mergeValidationResults(...results);
};

// URLのバリデーション
export const validateUrl = (url: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!url) {
    errors.push('URLは必須です');
  } else {
    try {
      new URL(url);
    } catch {
      errors.push('有効なURLを入力してください');
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 日付のバリデーション
export const validateDate = (dateString: string): ValidationResult => {
  const errors: string[] = [];
  
  if (!dateString) {
    errors.push('日付は必須です');
  } else {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      errors.push('有効な日付を入力してください');
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};
