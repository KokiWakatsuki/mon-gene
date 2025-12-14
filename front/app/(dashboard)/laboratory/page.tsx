'use client';

import { useState, useEffect, useRef } from 'react';
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

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  model?: string;
  api?: string;
  files?: ChatFile[];
}

interface ChatFile {
  name: string;
  type: string;
  data: string; // base64 encoded
  mimeType: string;
}

const API_OPTIONS = [
  { value: 'chatgpt', label: 'ChatGPT' },
  { value: 'claude', label: 'Claude' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'laboratory', label: 'Laboratory' }
];

const MODEL_OPTIONS = {
  chatgpt: [
    { value: 'gpt-5-pro', label: 'gpt-5-pro: GPT-5 Pro (次世代高性能モデル)' },
    { value: 'gpt-4o', label: 'gpt-4o: GPT-4 Omni (マルチモーダル高性能)' },
    { value: 'gpt-4o-2024-05-13', label: 'gpt-4o-2024-05-13: GPT-4o (2024年5月版)' },
    { value: 'gpt-3.5-turbo', label: 'gpt-3.5-turbo: GPT-3.5 Turbo (高速・コスト効率)' },
    { value: 'gpt-3.5-turbo-instruct', label: 'gpt-3.5-turbo-instruct: GPT-3.5 Instruct (指示特化)' }
  ],
  claude: [
    { value: 'claude-sonnet-4-5-20250929', label: 'claude-sonnet-4-5-20250929: Claude Sonnet 4.5 (自律コーディング・複雑エージェント)' },
    { value: 'claude-opus-4-1-20250805', label: 'claude-opus-4-1-20250805: Claude Opus 4.1 (専門的複雑タスク)' },
    { value: 'claude-sonnet-4-20250514', label: 'claude-sonnet-4-20250514: Claude Sonnet 4 (知能と速度のバランス)' },
    { value: 'claude-3-5-haiku', label: 'claude-3-5-haiku: Claude 3.5 Haiku (高速・大量処理)' }
  ],
  gemini: [
    { value: 'gemini-2.5-pro', label: 'gemini-2.5-pro: Gemini 2.5 Pro (最先端の思考・推論)' },
    { value: 'gemini-2.5-flash', label: 'gemini-2.5-flash: Gemini 2.5 Flash (価格・性能最適)' },
    { value: 'gemini-2.5-flash-lite', label: 'gemini-2.5-flash-lite: Gemini 2.5 Flash-Lite (高スループット)' },
    { value: 'gemini-2.5-pro-preview-03-25', label: 'gemini-2.5-pro-preview-03-25: Gemini 2.5 Pro Preview (3月版)' }
  ],
  laboratory: [
    { value: 'claude-sonnet-4-20250514', label: 'claude-sonnet-4-20250514: Claude Sonnet 4 (Laboratory専用)' }
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

  // Chat related states
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [showChat, setShowChat] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  // File upload states
  const [selectedFiles, setSelectedFiles] = useState<ChatFile[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetchUserData();
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [chatMessages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    Array.from(files).forEach(file => {
      // ファイルサイズ制限 (10MB)
      if (file.size > 10 * 1024 * 1024) {
        alert(`ファイル "${file.name}" が大きすぎます。10MB以下のファイルを選択してください。`);
        return;
      }

      const reader = new FileReader();
      reader.onload = (e) => {
        const result = e.target?.result as string;
        const base64Data = result.split(',')[1]; // Remove data URL prefix

        const fileType = file.type.startsWith('image/') ? 'image' : 'document';

        const newFile: ChatFile = {
          name: file.name,
          type: fileType,
          data: base64Data,
          mimeType: file.type,
        };

        setSelectedFiles(prev => [...prev, newFile]);
      };

      reader.readAsDataURL(file);
    });

    // Reset input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const removeFile = (index: number) => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== index));
  };

  const sendMessage = async () => {
    if ((!inputMessage.trim() && selectedFiles.length === 0) || isSending || !user) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: inputMessage.trim() || '(ファイルのみ)',
      timestamp: new Date(),
      files: selectedFiles.length > 0 ? selectedFiles : undefined
    };

    setChatMessages(prev => [...prev, userMessage]);
    setInputMessage('');
    const filesToSend = [...selectedFiles];
    setSelectedFiles([]);
    setIsSending(true);

    try {
      const token = localStorage.getItem('token');
      const requestBody: any = {
        message: userMessage.content
      };

      if (filesToSend.length > 0) {
        requestBody.files = filesToSend.map(file => ({
          name: file.name,
          type: file.type,
          data: file.data,
          mimeType: file.mimeType
        }));
      }

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(requestBody)
      });

      if (response.ok) {
        const data = await response.json();
        const assistantMessage: ChatMessage = {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: data.reply,
          timestamp: new Date(),
          model: data.model,
          api: data.api
        };
        setChatMessages(prev => [...prev, assistantMessage]);
      } else {
        const errorData = await response.text();
        const errorMessage: ChatMessage = {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: `エラー: ${errorData}`,
          timestamp: new Date()
        };
        setChatMessages(prev => [...prev, errorMessage]);
      }
    } catch (error) {
      console.error('メッセージ送信に失敗しました:', error);
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'メッセージの送信に失敗しました。ネットワーク接続を確認してください。',
        timestamp: new Date()
      };
      setChatMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsSending(false);
    }
  };

  const clearChat = () => {
    setChatMessages([]);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

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
              <h1 className="text-2xl font-bold text-gray-800">モデル実証実験画面</h1>
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
                  <div>
                    <span className="font-medium">設定AI:</span> {user.preferred_api || '未設定'}
                  </div>
                  <div>
                    <span className="font-medium">設定モデル:</span> {user.preferred_model || '未設定'}
                  </div>
                </div>
              </div>

              {/* API選択 */}
              <div>
                <h2 className="text-lg font-semibold mb-3">API選択</h2>
                {user.role !== 'admin' && (
                  <div className="mb-3 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                    <p className="text-sm text-yellow-800">
                      ⚠️ あなたの権限では Laboratory API のみ使用できます。
                    </p>
                  </div>
                )}
                <div className="space-y-2">
                  {API_OPTIONS.filter(api => user.role === 'admin' || api.value === 'laboratory').map((api) => (
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

              {/* AI チャット機能 */}
              <div className="border-t pt-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-lg font-semibold">AIチャット</h2>
                  <button
                    onClick={() => setShowChat(!showChat)}
                    className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
                  >
                    {showChat ? 'チャットを閉じる' : 'チャットを開く'}
                  </button>
                </div>

                {user.preferred_api && user.preferred_model ? (
                  <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
                    <p className="text-sm text-blue-800">
                      💡 現在の設定: {user.preferred_api} - {user.preferred_model}
                    </p>
                  </div>
                ) : (
                  <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                    <p className="text-sm text-yellow-800">
                      ⚠️ AIを使用するには、まずAPI設定とモデル設定を保存してください。
                    </p>
                  </div>
                )}

                {showChat && (
                  <div className="space-y-4">
                    {/* チャット履歴 */}
                    <div className="border border-gray-200 rounded-md p-4 h-96 overflow-y-auto bg-gray-50">
                      {chatMessages.length === 0 ? (
                        <div className="flex items-center justify-center h-full text-gray-500">
                          <p>チャットを開始してください</p>
                        </div>
                      ) : (
                        <div className="space-y-4">
                          {chatMessages.map((message) => (
                            <div
                              key={message.id}
                              className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
                            >
                              <div
                                className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                                  message.role === 'user'
                                    ? 'bg-blue-500 text-white'
                                    : 'bg-white border border-gray-200'
                                }`}
                              >
                                {/* ファイル表示 */}
                                {message.files && message.files.length > 0 && (
                                  <div className="mb-2">
                                    {message.files.map((file, fileIndex) => (
                                      <div key={fileIndex} className={`text-xs px-2 py-1 rounded mb-1 ${
                                        message.role === 'user' ? 'bg-blue-400' : 'bg-gray-100'
                                      }`}>
                                        <span>
                                          {file.type === 'image' ? '🖼️' : '📄'} {file.name}
                                        </span>
                                      </div>
                                    ))}
                                  </div>
                                )}
                                <p className="whitespace-pre-wrap">{message.content}</p>
                                <div className="text-xs mt-1 opacity-70">
                                  {message.timestamp.toLocaleTimeString()}
                                  {message.model && (
                                    <span className="ml-2">({message.api} - {message.model})</span>
                                  )}
                                </div>
                              </div>
                            </div>
                          ))}
                          {isSending && (
                            <div className="flex justify-start">
                              <div className="bg-white border border-gray-200 max-w-xs lg:max-w-md px-4 py-2 rounded-lg">
                                <p className="text-gray-500">入力中...</p>
                              </div>
                            </div>
                          )}
                          <div ref={messagesEndRef} />
                        </div>
                      )}
                    </div>

                    {/* ファイル選択表示 */}
                    {selectedFiles.length > 0 && (
                      <div className="border border-gray-200 rounded-md p-3 bg-gray-50">
                        <h4 className="text-sm font-medium mb-2">選択したファイル:</h4>
                        <div className="space-y-2">
                          {selectedFiles.map((file, index) => (
                            <div key={index} className="flex items-center justify-between bg-white px-3 py-2 rounded border">
                              <div className="flex items-center space-x-2">
                                <span className="text-sm">
                                  {file.type === 'image' ? '🖼️' : '📄'} {file.name}
                                </span>
                                <span className="text-xs text-gray-500">({file.mimeType})</span>
                              </div>
                              <button
                                onClick={() => removeFile(index)}
                                className="text-red-500 hover:text-red-700 text-sm"
                              >
                                ✕
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* メッセージ入力 */}
                    <div className="space-y-2">
                      <div className="flex space-x-2">
                        <div className="flex-1">
                          <textarea
                            value={inputMessage}
                            onChange={(e) => setInputMessage(e.target.value)}
                            onKeyPress={handleKeyPress}
                            placeholder="メッセージを入力してください... (Enterで送信、Shift+Enterで改行)"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                            rows={2}
                            disabled={isSending || !user.preferred_api || !user.preferred_model}
                          />
                        </div>
                        <div className="flex flex-col space-y-2">
                          <button
                            onClick={sendMessage}
                            disabled={(!inputMessage.trim() && selectedFiles.length === 0) || isSending || !user.preferred_api || !user.preferred_model}
                            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
                          >
                            {isSending ? '送信中...' : '送信'}
                          </button>
                          {chatMessages.length > 0 && (
                            <button
                              onClick={clearChat}
                              className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 text-sm"
                            >
                              クリア
                            </button>
                          )}
                        </div>
                      </div>

                      {/* ファイルアップロードボタン */}
                      <div className="flex space-x-2">
                        <input
                          type="file"
                          ref={fileInputRef}
                          onChange={handleFileSelect}
                          multiple
                          accept="image/*,.pdf,.txt,.doc,.docx,.csv,.json"
                          className="hidden"
                          disabled={isSending}
                        />
                        <button
                          onClick={() => fileInputRef.current?.click()}
                          disabled={isSending || !user.preferred_api || !user.preferred_model}
                          className="flex items-center space-x-2 bg-gray-500 text-white px-3 py-2 rounded hover:bg-gray-600 disabled:bg-gray-400 disabled:cursor-not-allowed text-sm"
                        >
                          <span>📎</span>
                          <span>ファイル添付</span>
                        </button>
                        <div className="text-xs text-gray-500 flex items-center">
                          画像、PDF、テキストファイル等（10MB以下）
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
