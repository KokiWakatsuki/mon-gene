from pydantic import BaseModel
from typing import Dict, Any, Optional

class GeometryDrawRequest(BaseModel):
    shape_type: str
    parameters: Dict[str, Any]
    labels: Optional[Dict[str, str]] = None

class GeometryResponse(BaseModel):
    success: bool
    image_base64: Optional[str] = None
    shape_type: Optional[str] = None
    error: Optional[str] = None

class CustomGeometryRequest(BaseModel):
    python_code: str
    problem_text: str

class CustomGeometryResponse(BaseModel):
    success: bool
    image_base64: Optional[str] = None
    problem_text: Optional[str] = None
    error: Optional[str] = None
