# 実装ガイドライン

## 移行戦略

### フェーズ1: 基盤整備（1-2週間）
1. **ディレクトリ構造の作成**
   - 新しいディレクトリ構造を段階的に作成
   - 既存コードを新構造に移動

2. **設定管理の統一**
   - 環境変数の整理と統一
   - 設定ファイルの分離

### フェーズ2: backコンテナのリファクタリング（2-3週間）
1. **レイヤードアーキテクチャの導入**
2. **責任の分離**
3. **テストの追加**

### フェーズ3: frontコンテナの整理（1-2週間）
1. **コンポーネントの再構成**
2. **カスタムフックの導入**
3. **API クライアントの整理**

### フェーズ4: coreコンテナの整理（1週間）
1. **モジュール分離**
2. **サービス層の導入**

## 各コンテナの実装詳細

### frontコンテナ

#### 1. API クライアントの実装例

```typescript
// lib/api/client.ts
class ApiClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  setToken(token: string) {
    this.token = token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const headers = {
      'Content-Type': 'application/json',
      ...(this.token && { Authorization: `Bearer ${this.token}` }),
      ...options.headers,
    };

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    return response.json();
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }
}

export const apiClient = new ApiClient(process.env.NEXT_PUBLIC_API_URL!);
```

#### 2. カスタムフックの実装例

```typescript
// hooks/useAuth.ts
export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const login = async (credentials: LoginCredentials) => {
    const response = await authApi.login(credentials);
    setUser(response.user);
    apiClient.setToken(response.token);
    localStorage.setItem('token', response.token);
  };

  const logout = () => {
    setUser(null);
    apiClient.setToken('');
    localStorage.removeItem('token');
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      apiClient.setToken(token);
      // トークンの検証とユーザー情報の取得
    }
    setLoading(false);
  }, []);

  return { user, loading, login, logout };
}
```

### backコンテナ

#### 1. レイヤードアーキテクチャの実装例

```go
// internal/api/handlers/problems.go
type ProblemHandler struct {
    problemService services.ProblemService
}

func NewProblemHandler(problemService services.ProblemService) *ProblemHandler {
    return &ProblemHandler{
        problemService: problemService,
    }
}

func (h *ProblemHandler) GenerateProblem(w http.ResponseWriter, r *http.Request) {
    var req models.GenerateProblemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request")
        return
    }

    problem, err := h.problemService.GenerateProblem(r.Context(), req)
    if err != nil {
        utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
        return
    }

    utils.WriteJSONResponse(w, http.StatusOK, problem)
}
```

```go
// internal/services/problem_service.go
type ProblemService interface {
    GenerateProblem(ctx context.Context, req models.GenerateProblemRequest) (*models.Problem, error)
}

type problemService struct {
    claudeClient clients.ClaudeClient
    coreClient   clients.CoreClient
    problemRepo  repositories.ProblemRepository
}

func NewProblemService(
    claudeClient clients.ClaudeClient,
    coreClient clients.CoreClient,
    problemRepo repositories.ProblemRepository,
) ProblemService {
    return &problemService{
        claudeClient: claudeClient,
        coreClient:   coreClient,
        problemRepo:  problemRepo,
    }
}

func (s *problemService) GenerateProblem(ctx context.Context, req models.GenerateProblemRequest) (*models.Problem, error) {
    // 1. Claude APIで問題文を生成
    content, err := s.claudeClient.GenerateContent(ctx, req.Prompt)
    if err != nil {
        return nil, fmt.Errorf("failed to generate content: %w", err)
    }

    // 2. 図形が必要かどうかを分析
    analysis, err := s.coreClient.AnalyzeProblem(ctx, content, req.Filters)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze problem: %w", err)
    }

    // 3. 必要に応じて図形を生成
    var imageBase64 string
    if analysis.NeedsGeometry {
        imageBase64, err = s.coreClient.GenerateGeometry(ctx, analysis.DetectedShapes[0], analysis.SuggestedParameters)
        if err != nil {
            return nil, fmt.Errorf("failed to generate geometry: %w", err)
        }
    }

    // 4. 問題をデータベースに保存
    problem := &models.Problem{
        Content:     content,
        ImageBase64: imageBase64,
        Subject:     req.Subject,
        Filters:     req.Filters,
        CreatedAt:   time.Now(),
    }

    if err := s.problemRepo.Create(ctx, problem); err != nil {
        return nil, fmt.Errorf("failed to save problem: %w", err)
    }

    return problem, nil
}
```

#### 2. 依存性注入の実装例

```go
// cmd/server/main.go
func main() {
    cfg := config.Load()
    
    // データベース接続
    db, err := config.NewDatabase(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // リポジトリの初期化
    problemRepo := repositories.NewProblemRepository(db)
    userRepo := repositories.NewUserRepository(db)

    // クライアントの初期化
    claudeClient := clients.NewClaudeClient(cfg.ClaudeAPIKey)
    coreClient := clients.NewCoreClient(cfg.CoreAPIURL)

    // サービスの初期化
    problemService := services.NewProblemService(claudeClient, coreClient, problemRepo)
    authService := services.NewAuthService(userRepo)

    // ハンドラーの初期化
    problemHandler := handlers.NewProblemHandler(problemService)
    authHandler := handlers.NewAuthHandler(authService)

    // ルーターの設定
    router := routes.NewRouter(problemHandler, authHandler)

    // サーバー起動
    log.Printf("Server starting on port %s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
```

### coreコンテナ

#### 1. サービス層の実装例

```python
# app/services/geometry_service.py
from typing import Optional, Dict, Any
from app.core.geometry.renderer import GeometryRenderer
from app.core.geometry.analyzer import GeometryAnalyzer
from app.models.geometry import GeometryRequest, GeometryResponse

class GeometryService:
    def __init__(self):
        self.renderer = GeometryRenderer()
        self.analyzer = GeometryAnalyzer()

    async def generate_geometry(
        self, 
        shape_type: str, 
        parameters: Dict[str, Any],
        labels: Optional[Dict[str, str]] = None
    ) -> GeometryResponse:
        """図形を生成する"""
        try:
            # 図形を描画
            figure = self.renderer.render_shape(shape_type, parameters, labels)
            
            # Base64エンコード
            image_base64 = self.renderer.figure_to_base64(figure)
            
            return GeometryResponse(
                success=True,
                image_base64=image_base64,
                shape_type=shape_type
            )
        except Exception as e:
            return GeometryResponse(
                success=False,
                error=str(e)
            )

    async def analyze_problem(
        self, 
        problem_text: str, 
        unit_parameters: Dict[str, Any]
    ) -> Dict[str, Any]:
        """問題文を解析して図形の必要性を判定"""
        return self.analyzer.analyze_geometry_needs(problem_text, unit_parameters)
```

#### 2. 図形描画エンジンの実装例

```python
# app/core/geometry/renderer.py
import matplotlib.pyplot as plt
from typing import Dict, Any, Optional
from app.core.geometry.shapes.basic_shapes import BasicShapeRenderer
from app.core.geometry.shapes.solid_shapes import SolidShapeRenderer

class GeometryRenderer:
    def __init__(self):
        self.basic_renderer = BasicShapeRenderer()
        self.solid_renderer = SolidShapeRenderer()

    def render_shape(
        self, 
        shape_type: str, 
        parameters: Dict[str, Any],
        labels: Optional[Dict[str, str]] = None
    ) -> plt.Figure:
        """図形タイプに応じて適切なレンダラーを選択"""
        
        if shape_type in ['triangle', 'circle', 'square', 'rectangle']:
            return self.basic_renderer.render(shape_type, parameters, labels)
        elif shape_type in ['cube', 'sphere', 'cone', 'pyramid', 'cylinder']:
            return self.solid_renderer.render(shape_type, parameters, labels)
        else:
            raise ValueError(f"Unsupported shape type: {shape_type}")

    def figure_to_base64(self, figure: plt.Figure) -> str:
        """matplotlib図をbase64文字列に変換"""
        import io
        import base64
        
        buffer = io.BytesIO()
        figure.savefig(buffer, format='png', dpi=150, bbox_inches='tight')
        buffer.seek(0)
        image_base64 = base64.b64encode(buffer.getvalue()).decode()
        plt.close(figure)
        return image_base64
```

## データベース設計

### テーブル設計例

```sql
-- users テーブル
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    school_code VARCHAR(10) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- problems テーブル
CREATE TABLE problems (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    subject VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    image_base64 LONGTEXT,
    filters JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- sessions テーブル
CREATE TABLE sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## テスト戦略

### 1. フロントエンドテスト

```typescript
// __tests__/components/ProblemCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import ProblemCard from '@/components/features/problems/ProblemCard';

describe('ProblemCard', () => {
  const mockProps = {
    id: '1',
    title: 'Test Problem',
    content: 'Test content',
    onPreview: jest.fn(),
    onPrint: jest.fn(),
  };

  it('renders problem card correctly', () => {
    render(<ProblemCard {...mockProps} />);
    expect(screen.getByText('Test Problem')).toBeInTheDocument();
    expect(screen.getByText('Test content')).toBeInTheDocument();
  });

  it('calls onPreview when preview button is clicked', () => {
    render(<ProblemCard {...mockProps} />);
    fireEvent.click(screen.getByText('プレビュー'));
    expect(mockProps.onPreview).toHaveBeenCalledWith('1');
  });
});
```

### 2. バックエンドテスト

```go
// internal/services/problem_service_test.go
func TestProblemService_GenerateProblem(t *testing.T) {
    // モックの設定
    mockClaudeClient := &mocks.ClaudeClient{}
    mockCoreClient := &mocks.CoreClient{}
    mockProblemRepo := &mocks.ProblemRepository{}

    service := services.NewProblemService(mockClaudeClient, mockCoreClient, mockProblemRepo)

    // テストケース
    req := models.GenerateProblemRequest{
        Prompt:  "数学の問題を生成してください",
        Subject: "数学",
        Filters: map[string]interface{}{"difficulty": "medium"},
    }

    // モックの期待値設定
    mockClaudeClient.On("GenerateContent", mock.Anything, req.Prompt).Return("生成された問題", nil)
    mockCoreClient.On("AnalyzeProblem", mock.Anything, mock.Anything, mock.Anything).Return(&models.AnalysisResponse{
        NeedsGeometry: false,
    }, nil)
    mockProblemRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // テスト実行
    problem, err := service.GenerateProblem(context.Background(), req)

    // アサーション
    assert.NoError(t, err)
    assert.NotNil(t, problem)
    assert.Equal(t, "生成された問題", problem.Content)
    mockClaudeClient.AssertExpectations(t)
    mockCoreClient.AssertExpectations(t)
    mockProblemRepo.AssertExpectations(t)
}
```

### 3. コアサービステスト

```python
# tests/test_geometry_service.py
import pytest
from app.services.geometry_service import GeometryService

@pytest.fixture
def geometry_service():
    return GeometryService()

@pytest.mark.asyncio
async def test_generate_geometry_success(geometry_service):
    # テストデータ
    shape_type = "square"
    parameters = {"side_length": 5}
    
    # テスト実行
    result = await geometry_service.generate_geometry(shape_type, parameters)
    
    # アサーション
    assert result.success is True
    assert result.image_base64 is not None
    assert result.shape_type == shape_type

@pytest.mark.asyncio
async def test_generate_geometry_invalid_shape(geometry_service):
    # テストデータ
    shape_type = "invalid_shape"
    parameters = {}
    
    # テスト実行
    result = await geometry_service.generate_geometry(shape_type, parameters)
    
    # アサーション
    assert result.success is False
    assert result.error is not None
```

## デプロイメント

### Docker Compose設定の改善

```yaml
# docker-compose.yml
version: "3.8"

services:
  front:
    build:
      context: ./front
      dockerfile: ../docker/Dockerfile.front
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
      - NEXT_PUBLIC_USE_REAL_API=true
    depends_on:
      - back
    volumes:
      - ./front:/app
      - /app/node_modules
    networks:
      - mongene-network

  back:
    build:
      context: ./back
      dockerfile: ../docker/Dockerfile.back
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=mysql://user:password@db:3306/mongene
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - CORE_API_URL=http://core:1234
    depends_on:
      - db
      - core
    volumes:
      - ./back:/app
    networks:
      - mongene-network

  core:
    build:
      context: ./core
      dockerfile: ../docker/Dockerfile.core
    ports:
      - "1234:1234"
    volumes:
      - ./core:/app
    networks:
      - mongene-network

  db:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=rootpass
      - MYSQL_DATABASE=mongene
      - MYSQL_USER=user
      - MYSQL_PASSWORD=password
    volumes:
      - db-data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - mongene-network

volumes:
  db-data:

networks:
  mongene-network:
    driver: bridge
```

## 移行チェックリスト

### フェーズ1: 基盤整備
- [ ] 新しいディレクトリ構造の作成
- [ ] 環境変数の整理
- [ ] Docker設定の更新
- [ ] データベーススキーマの設計

### フェーズ2: backコンテナリファクタリング
- [ ] レイヤードアーキテクチャの実装
- [ ] ハンドラーの分離
- [ ] サービス層の実装
- [ ] リポジトリ層の実装
- [ ] クライアント層の実装
- [ ] ユニットテストの追加

### フェーズ3: frontコンテナ整理
- [ ] コンポーネントの再構成
- [ ] カスタムフックの実装
- [ ] API クライアントの整理
- [ ] 型定義の整理
- [ ] テストの追加

### フェーズ4: coreコンテナ整理
- [ ] サービス層の実装
- [ ] モジュール分離
- [ ] テストの追加

### フェーズ5: 統合テスト・デプロイ
- [ ] 統合テストの実装
- [ ] CI/CDパイプラインの設定
- [ ] 本番環境への適用

この実装ガイドラインに従って段階的に移行を進めることで、システムの保守性と拡張性を大幅に向上させることができます。
