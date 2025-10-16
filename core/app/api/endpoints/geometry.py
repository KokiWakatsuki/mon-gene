from fastapi import APIRouter, HTTPException
from app.models.geometry import GeometryDrawRequest, GeometryResponse
from app.services.geometry_service import GeometryService

router = APIRouter()
geometry_service = GeometryService()

@router.post("/draw-geometry", response_model=GeometryResponse)
async def draw_geometry(request: GeometryDrawRequest):
    """図形描画エンドポイント"""
    try:
        result = await geometry_service.generate_geometry(
            request.shape_type,
            request.parameters,
            request.labels
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/draw-custom-geometry")
async def draw_custom_geometry(request: dict):
    """カスタムPythonコードを実行して図形を描画"""
    try:
        python_code = request.get("python_code", "")
        problem_text = request.get("problem_text", "")
        
        result = await geometry_service.generate_custom_geometry(python_code, problem_text)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/execute-python")
async def execute_python(request: dict):
    """Pythonコードを実行して結果を返す（数値計算用）"""
    try:
        python_code = request.get("python_code", "")
        
        if not python_code:
            raise HTTPException(status_code=400, detail="python_code is required")
        
        result = await geometry_service.execute_python_code(python_code)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
