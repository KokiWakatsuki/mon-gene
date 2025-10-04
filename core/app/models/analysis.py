from pydantic import BaseModel
from typing import Dict, Any, List, Optional

class ProblemAnalysisRequest(BaseModel):
    problem_text: str
    unit_parameters: Dict[str, Any]
    subject: str = "math"

class ProblemAnalysisResponse(BaseModel):
    success: bool
    needs_geometry: bool
    detected_shapes: List[str]
    suggested_parameters: Dict[str, Dict[str, Any]]
    error: Optional[str] = None
