from fastapi import APIRouter, HTTPException
from app.models.analysis import ProblemAnalysisRequest, ProblemAnalysisResponse
from app.services.analysis_service import AnalysisService

router = APIRouter()
analysis_service = AnalysisService()

@router.post("/analyze-problem", response_model=ProblemAnalysisResponse)
async def analyze_problem(request: ProblemAnalysisRequest):
    """問題を分析して図形の必要性を判定"""
    try:
        result = await analysis_service.analyze_problem(
            request.problem_text,
            request.unit_parameters,
            request.subject
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
