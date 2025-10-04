from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api.endpoints import geometry, analysis, pdf

# FastAPIアプリケーションの初期化
app = FastAPI(
    title="Mongene Core API",
    description="AI処理、図形生成、PDF生成を担当するコアサービス",
    version="1.0.0"
)

# CORS設定
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ルーターの登録
app.include_router(geometry.router, tags=["geometry"])
app.include_router(analysis.router, tags=["analysis"])
app.include_router(pdf.router, tags=["pdf"])

@app.get("/")
async def root():
    return {
        "message": "Mongene Core API Server is running",
        "service": "mongene-core",
        "version": "1.0.0"
    }

@app.get("/health")
async def health():
    return {
        "status": "ok",
        "message": "Mongene Core API Server is running",
        "service": "mongene-core",
        "version": "1.0.0"
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=1234)
