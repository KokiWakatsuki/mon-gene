from fastapi import APIRouter, HTTPException
from app.models.pdf import PDFGenerateRequest, PDFGenerateResponse
from app.services.pdf_service import PDFService

router = APIRouter()
pdf_service = PDFService()

@router.post("/generate-pdf", response_model=PDFGenerateResponse)
async def generate_pdf(request: PDFGenerateRequest):
    """PDF生成エンドポイント"""
    try:
        result = await pdf_service.generate_pdf(
            request.problem_text,
            request.image_base64,
            request.solution_text
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
