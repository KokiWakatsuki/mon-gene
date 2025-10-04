from pydantic import BaseModel
from typing import Optional

class PDFGenerateRequest(BaseModel):
    problem_text: str
    image_base64: Optional[str] = None
    solution_text: Optional[str] = None

class PDFGenerateResponse(BaseModel):
    success: bool
    pdf_base64: Optional[str] = None
    error: Optional[str] = None
