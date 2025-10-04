import matplotlib
matplotlib.use('Agg')  # GUI不要のバックエンドを使用
import matplotlib.pyplot as plt
import matplotlib.patches as patches
import numpy as np
import base64
import io
from mpl_toolkits.mplot3d import Axes3D
from mpl_toolkits.mplot3d.art3d import Poly3DCollection
from mpl_toolkits import mplot3d

from app.models.geometry import GeometryResponse, CustomGeometryResponse

class GeometryService:
    """図形生成サービス"""
    
    async def generate_geometry(self, shape_type: str, parameters: dict, labels: dict = None) -> GeometryResponse:
        """図形を生成する"""
        try:
            fig, ax = plt.subplots(1, 1, figsize=(8, 6))
            
            if shape_type == "triangle":
                self._draw_triangle(ax, parameters, labels)
            elif shape_type == "rectangle":
                self._draw_rectangle(ax, parameters, labels)
            elif shape_type == "circle":
                self._draw_circle(ax, parameters, labels)
            elif shape_type == "square":
                self._draw_square(ax, parameters, labels)
            else:
                # デフォルトは正方形
                self._draw_square(ax, {"side": 5}, labels)
            
            ax.set_aspect('equal')
            ax.grid(True, alpha=0.3)
            plt.tight_layout()
            
            # 画像をBase64エンコード
            buffer = io.BytesIO()
            plt.savefig(buffer, format='png', dpi=150, bbox_inches='tight')
            buffer.seek(0)
            image_base64 = base64.b64encode(buffer.getvalue()).decode()
            plt.close()
            
            return GeometryResponse(
                success=True,
                image_base64=image_base64,
                shape_type=shape_type
            )
        except Exception as e:
            plt.close()
            return GeometryResponse(
                success=False,
                image_base64="",
                shape_type=shape_type
            )
    
    async def generate_custom_geometry(self, python_code: str, problem_text: str) -> CustomGeometryResponse:
        """カスタムPythonコードで図形を生成する"""
        print(f"🔍 generate_custom_geometry called")
        print(f"🔍 python_code length: {len(python_code)}")
        print(f"🔍 problem_text length: {len(problem_text)}")
        
        try:
            # 安全な実行環境を準備
            safe_globals = {
                'plt': plt,
                'patches': patches,
                'np': np,
                'numpy': np,
                'Axes3D': Axes3D,
                'Poly3DCollection': Poly3DCollection,
                'matplotlib': matplotlib,
                'mplot3d': mplot3d,
                'io': io,
                'base64': base64,
                'Polygon': patches.Polygon,  # matplotlib.patches.Polygon のエイリアス
                '__builtins__': {
                    '__import__': __import__,
                    'len': len,
                    'range': range,
                    'enumerate': enumerate,
                    'zip': zip,
                    'map': map,  # map関数を追加
                    'filter': filter,  # filter関数も追加
                    'list': list,
                    'dict': dict,
                    'tuple': tuple,
                    'set': set,
                    'str': str,
                    'int': int,
                    'float': float,
                    'bool': bool,
                    'min': min,
                    'max': max,
                    'abs': abs,
                    'round': round,
                    'sum': sum,
                    'print': print,
                    'ord': ord,
                    'chr': chr,
                }
            }
            
            print(f"🔍 About to execute Python code")
            print(f"🔍 Python code preview: {python_code[:200]}...")
            
            # Pythonコードを実行
            exec(python_code, safe_globals)
            print(f"🔍 Python code executed successfully")
            
            # 画像をBase64エンコード
            buffer = io.BytesIO()
            print(f"🔍 About to save figure")
            plt.savefig(buffer, format='png', dpi=150, bbox_inches='tight')
            buffer.seek(0)
            image_data = buffer.getvalue()
            print(f"🔍 Image data length: {len(image_data)}")
            
            if len(image_data) == 0:
                print(f"❌ Image data is empty!")
                plt.close()
                return CustomGeometryResponse(
                    success=False,
                    image_base64="",
                    problem_text=problem_text,
                    error="Generated image data is empty"
                )
            
            image_base64 = base64.b64encode(image_data).decode()
            print(f"🔍 Base64 encoded image length: {len(image_base64)}")
            plt.close()
            
            return CustomGeometryResponse(
                success=True,
                image_base64=image_base64,
                problem_text=problem_text
            )
        except Exception as e:
            print(f"❌ Error in generate_custom_geometry: {str(e)}")
            print(f"❌ Error type: {type(e).__name__}")
            import traceback
            print(f"❌ Traceback: {traceback.format_exc()}")
            plt.close()
            return CustomGeometryResponse(
                success=False,
                image_base64="",
                problem_text=problem_text,
                error=str(e)
            )
    
    def _draw_triangle(self, ax, parameters, labels):
        """三角形を描画"""
        width = parameters.get("width", 5)
        height = parameters.get("height", 4)
        
        # 三角形の頂点
        triangle = patches.Polygon([(0, 0), (width, 0), (width/2, height)], 
                                 closed=True, fill=False, edgecolor='blue', linewidth=2)
        ax.add_patch(triangle)
        
        # ラベル
        if labels:
            ax.text(0, -0.3, 'A', fontsize=12, ha='center')
            ax.text(width, -0.3, 'B', fontsize=12, ha='center')
            ax.text(width/2, height+0.2, 'C', fontsize=12, ha='center')
        
        ax.set_xlim(-1, width+1)
        ax.set_ylim(-1, height+1)
    
    def _draw_rectangle(self, ax, parameters, labels):
        """長方形を描画"""
        width = parameters.get("width", 6)
        height = parameters.get("height", 4)
        
        rectangle = patches.Rectangle((0, 0), width, height, 
                                    fill=False, edgecolor='blue', linewidth=2)
        ax.add_patch(rectangle)
        
        # ラベル
        if labels:
            ax.text(0, -0.3, 'A', fontsize=12, ha='center')
            ax.text(width, -0.3, 'B', fontsize=12, ha='center')
            ax.text(width, height+0.2, 'C', fontsize=12, ha='center')
            ax.text(0, height+0.2, 'D', fontsize=12, ha='center')
        
        ax.set_xlim(-1, width+1)
        ax.set_ylim(-1, height+1)
    
    def _draw_square(self, ax, parameters, labels):
        """正方形を描画"""
        side = parameters.get("side", 5)
        
        square = patches.Rectangle((0, 0), side, side, 
                                 fill=False, edgecolor='blue', linewidth=2)
        ax.add_patch(square)
        
        # ラベル
        if labels:
            ax.text(0, -0.3, 'A', fontsize=12, ha='center')
            ax.text(side, -0.3, 'B', fontsize=12, ha='center')
            ax.text(side, side+0.2, 'C', fontsize=12, ha='center')
            ax.text(0, side+0.2, 'D', fontsize=12, ha='center')
        
        ax.set_xlim(-1, side+1)
        ax.set_ylim(-1, side+1)
    
    def _draw_circle(self, ax, parameters, labels):
        """円を描画"""
        radius = parameters.get("radius", 3)
        
        circle = patches.Circle((0, 0), radius, 
                              fill=False, edgecolor='blue', linewidth=2)
        ax.add_patch(circle)
        
        # 中心点
        ax.plot(0, 0, 'ro', markersize=4)
        if labels:
            ax.text(0, -0.3, 'O', fontsize=12, ha='center')
        
        ax.set_xlim(-radius-1, radius+1)
        ax.set_ylim(-radius-1, radius+1)
