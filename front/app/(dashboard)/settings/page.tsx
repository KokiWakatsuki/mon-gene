'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface User {
  id: number;
  school_code: string;
  email: string;
  role: string;
  preferred_api: string;
  preferred_model: string;
  problem_generation_limit: number;
}

const API_OPTIONS = [
  { value: 'chatgpt', label: 'ChatGPT' },
  { value: 'claude', label: 'Claude' },
  { value: 'gemini', label: 'Gemini' }
];

const MODEL_OPTIONS = {
  chatgpt: [
    { value: 'gpt-4', label: 'GPT-4' },
    { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
    { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' }
  ],
  claude: [
    { value: 'claude-3-opus', label: 'Claude 3 Opus' },
    { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
    { value: 'claude-3-haiku', label: 'Claude 3 Haiku' }
  ],
  gemini: [
    { value: 'gemini-pro', label: 'Gemini Pro' },
    { value: 'gemini-pro-vision', label: 'Gemini Pro Vision' }
  ]
};

export default function SettingsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [selectedAPI, setSelectedAPI] = useState('');
  const [selectedModel, setSelectedModel] = useState('');
  const [customModel, setCustomModel] = useState('');
  const [useCustomModel, setUseCustomModel] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchUserData();
  }, []);

  const fetchUserData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        router.push('/login');
        return;
      }

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/profile`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const userData = await response.json();
        setUser(userData);
        setSelectedAPI(userData.preferred_api || 'claude');
        setSelectedModel(userData.preferred_model || 'claude-3-haiku');
        
        // カスタムモデルかどうかチェック
        const isCustom = !MODEL_OPTIONS[userData.preferred_api as keyof typeof MODEL_OPTIONS]?.some(
          model => model.value === userData.preferred_model
        );
        setUseCustomModel(isCustom);
        if (isCustom) {
          setCustomModel(userData.preferred_model);
        }
      } else {
        router.push('/login');
      }
    } catch (error) {
      console.error('ユーザーデータの取得に失敗しました:', error);
      router.push('/login');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!user) return;

    setSaving(true);
    try {
      const token = localStorage.getItem('token');
      const finalModel = useCustomModel ? customModel : selectedModel;

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/user/settings`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          preferred_api: selectedAPI,
          preferred_model: finalModel
        })
      });

      if (response.ok) {
        alert('設定を保存しました');
        fetchUserData(); // データを再取得
      } else {
        alert('設定の保存に失敗しました');
      }
    } catch (error) {
      console.error('設定の保存に失敗しました:', error);
      alert('設定の保存に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    router.push('/login');
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="text-lg">読み込み中...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  // adminユーザー以外はアクセス不可
  if (user.role !== 'admin') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="bg-white p-8 rounded-lg shadow-md">
          <h1 className="text-2xl font-bold text-red-600 mb-4">アクセス拒否</h1>
          <p className="text-gray-600 mb-4">この機能はadminユーザーのみ利用できます。</p>
          <button
            onClick={() => router.push('/problems')}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            問題一覧に戻る
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-2xl mx-auto">
          <div className="bg-white rounded-lg shadow-md p-6">
            <div className="flex justify-between items-center mb-6">
              <h1 className="text-2xl font-bold text-gray-800">設定</h1>
              <div className="space-x-2">
                <button
                  onClick={() => router.push('/problems')}
                  className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
                >
                  戻る
                </button>
                <button
                  onClick={handleLogout}
                  className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
                >
                  ログアウト
                </button>
              </div>
            </div>

            <div className="space-y-6">
              {/* ユーザー情報 */}
              <div className="border-b pb-4">
                <h2 className="text-lg font-semibold mb-2">ユーザー情報</h2>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="font-medium">学校コード:</span> {user.school_code}
                  </div>
                  <div>
                    <span className="font-medium">メール:</span> {user.email}
                  </div>
                  <div>
                    <span className="font-medium">権限:</span> {user.role}
                  </div>
                  <div>
                    <span className="font-medium">生成制限:</span> {user.problem_generation_limit === -1 ? '無制限' : user.problem_generation_limit}
                  </div>
                </div>
              </div>

              {/* API選択 */}
              <div>
                <h2 className="text-lg font-semibold mb-3">API選択</h2>
                <div className="space-y-2">
                  {API_OPTIONS.map((api) => (
                    <label key={api.value} className="flex items-center space-x-2">
                      <input
                        type="radio"
                        name="api"
                        value={api.value}
                        checked={selectedAPI === api.value}
                        onChange={(e) => {
                          setSelectedAPI(e.target.value);
                          // APIが変更されたら、そのAPIのデフォルトモデルを選択
                          const defaultModel = MODEL_OPTIONS[e.target.value as keyof typeof MODEL_OPTIONS][0].value;
                          setSelectedModel(defaultModel);
                          setUseCustomModel(false);
                          setCustomModel('');
                        }}
                        className="w-4 h-4 text-blue-600"
                      />
                      <span>{api.label}</span>
                    </label>
                  ))}
                </div>
              </div>

              {/* モデル選択 */}
              <div>
                <h2 className="text-lg font-semibold mb-3">モデル選択</h2>
                <div className="space-y-2">
                  {MODEL_OPTIONS[selectedAPI as keyof typeof MODEL_OPTIONS]?.map((model) => (
                    <label key={model.value} className="flex items-center space-x-2">
                      <input
                        type="radio"
                        name="model"
                        value={model.value}
                        checked={!useCustomModel && selectedModel === model.value}
                        onChange={(e) => {
                          setSelectedModel(e.target.value);
                          setUseCustomModel(false);
                          setCustomModel('');
                        }}
                        className="w-4 h-4 text-blue-600"
                      />
                      <span>{model.label}</span>
                    </label>
                  ))}
                  
                  {/* カスタムモデル */}
                  <label className="flex items-center space-x-2">
                    <input
                      type="radio"
                      name="model"
                      checked={useCustomModel}
                      onChange={() => {
                        setUseCustomModel(true);
                        setSelectedModel('');
                      }}
                      className="w-4 h-4 text-blue-600"
                    />
                    <span>その他（カスタム）</span>
                  </label>
                  
                  {useCustomModel && (
                    <div className="ml-6">
                      <input
                        type="text"
                        value={customModel}
                        onChange={(e) => setCustomModel(e.target.value)}
                        placeholder="モデル名を入力してください"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                  )}
                </div>
              </div>

              {/* 保存ボタン */}
              <div className="pt-4">
                <button
                  onClick={handleSave}
                  disabled={saving || (useCustomModel && !customModel.trim())}
                  className="w-full bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
                >
                  {saving ? '保存中...' : '設定を保存'}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
