import re
from app.models.analysis import ProblemAnalysisResponse

class AnalysisService:
    """問題解析サービス"""
    
    async def analyze_problem(self, problem_text: str, unit_parameters: dict, subject: str = "math") -> ProblemAnalysisResponse:
        """問題を解析して図形の必要性を判定する"""
        
        # キーワードベースの図形検出
        geometry_keywords = [
            "図形", "三角形", "四角形", "正方形", "長方形", "円", "楕円",
            "グラフ", "座標", "面積", "体積", "周囲", "直線", "曲線",
            "立方体", "直方体", "円錐", "球", "角度", "辺", "頂点",
            "対角線", "半径", "直径", "高さ", "底面", "側面",
            "三角柱", "四角柱", "円柱", "角柱", "柱", "錐", "pyramid"
        ]
        
        needs_geometry = any(keyword in problem_text for keyword in geometry_keywords)
        
        detected_shapes = []
        suggested_parameters = {}
        
        if needs_geometry:
            # 三角形の検出
            if any(word in problem_text for word in ["三角形", "triangle"]):
                detected_shapes.append("triangle")
                # 数値を抽出して適切なパラメータを設定
                numbers = self._extract_numbers(problem_text)
                if len(numbers) >= 2:
                    suggested_parameters["triangle"] = {"width": numbers[0], "height": numbers[1]}
                else:
                    suggested_parameters["triangle"] = {"width": 5, "height": 4}
            
            # 四角形・正方形・長方形の検出
            if any(word in problem_text for word in ["四角形", "正方形", "長方形", "rectangle", "square"]):
                if "正方形" in problem_text or "square" in problem_text:
                    detected_shapes.append("square")
                    numbers = self._extract_numbers(problem_text)
                    if numbers:
                        suggested_parameters["square"] = {"side": numbers[0]}
                    else:
                        suggested_parameters["square"] = {"side": 5}
                else:
                    detected_shapes.append("rectangle")
                    numbers = self._extract_numbers(problem_text)
                    if len(numbers) >= 2:
                        suggested_parameters["rectangle"] = {"width": numbers[0], "height": numbers[1]}
                    else:
                        suggested_parameters["rectangle"] = {"width": 6, "height": 4}
            
            # 円の検出
            if any(word in problem_text for word in ["円", "circle"]):
                detected_shapes.append("circle")
                numbers = self._extract_numbers(problem_text)
                if numbers:
                    # 半径または直径を検出
                    if "半径" in problem_text or "radius" in problem_text:
                        suggested_parameters["circle"] = {"radius": numbers[0]}
                    elif "直径" in problem_text or "diameter" in problem_text:
                        suggested_parameters["circle"] = {"radius": numbers[0] / 2}
                    else:
                        suggested_parameters["circle"] = {"radius": numbers[0]}
                else:
                    suggested_parameters["circle"] = {"radius": 3}
        
        return ProblemAnalysisResponse(
            success=True,
            needs_geometry=needs_geometry,
            detected_shapes=detected_shapes,
            suggested_parameters=suggested_parameters
        )
    
    def _extract_numbers(self, text: str) -> list:
        """テキストから数値を抽出する"""
        # 整数と小数を抽出
        numbers = re.findall(r'\d+\.?\d*', text)
        return [float(num) for num in numbers if float(num) > 0]
