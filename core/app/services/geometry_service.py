import matplotlib
matplotlib.use('Agg')  # GUI不要のバックエンドを使用
import matplotlib.pyplot as plt
import matplotlib.patches as patches
import matplotlib.font_manager as fm
import numpy as np
import base64
import io
from mpl_toolkits.mplot3d import Axes3D
from mpl_toolkits.mplot3d.art3d import Poly3DCollection
from mpl_toolkits import mplot3d

from app.models.geometry import GeometryResponse, CustomGeometryResponse

# 日本語フォントの設定
def setup_japanese_font():
    """日本語フォントを設定する"""
    try:
        # 利用可能な日本語フォントを探す
        japanese_fonts = [
            'Noto Sans CJK JP',
            'Noto Sans JP',
            'IPAexGothic',
            'IPAGothic',
            'TakaoPGothic',
            'VL Gothic',
            'DejaVu Sans'
        ]
        
        available_fonts = [f.name for f in fm.fontManager.ttflist]
        
        for font_name in japanese_fonts:
            if font_name in available_fonts:
                plt.rcParams['font.family'] = font_name
                print(f"✅ Japanese font set to: {font_name}")
                return font_name
        
        # フォールバック: システムのデフォルトフォントを使用
        print("⚠️ No Japanese font found, using default font")
        # マイナス記号の文字化けを防ぐ
        plt.rcParams['axes.unicode_minus'] = False
        return None
        
    except Exception as e:
        print(f"⚠️ Error setting up Japanese font: {e}")
        plt.rcParams['axes.unicode_minus'] = False
        return None

# 初期化時にフォントを設定
setup_japanese_font()

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
            elif shape_type == "cuboid":
                # 3D図形の場合は別の処理
                return await self._draw_cuboid_3d(parameters, labels)
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
            # 日本語フォントを再設定（exec実行前に毎回設定）
            font_name = setup_japanese_font()
            if font_name:
                print(f"✅ Japanese font configured for this request: {font_name}")
            else:
                print(f"⚠️ Using default font for this request")
            
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
    
    async def _draw_cuboid_3d(self, parameters, labels):
        """3D直方体を描画"""
        try:
            width = parameters.get("width", 6)
            depth = parameters.get("depth", 6)
            height = parameters.get("height", 8)
            
            # 3Dプロット
            fig = plt.figure(figsize=(10, 8))
            ax = fig.add_subplot(111, projection='3d')
            
            # 直方体の頂点を定義 (ABCD-EFGHの順番)
            vertices = np.array([
                [0, 0, 0],        # A (0,0,0)
                [width, 0, 0],    # B (width,0,0)
                [width, depth, 0], # C (width,depth,0)
                [0, depth, 0],    # D (0,depth,0)
                [0, 0, height],   # E (0,0,height)
                [width, 0, height], # F (width,0,height)
                [width, depth, height], # G (width,depth,height)
                [0, depth, height]  # H (0,depth,height)
            ])
            
            # 各面を定義
            faces = [
                [vertices[0], vertices[1], vertices[2], vertices[3]],  # 底面 ABCD
                [vertices[4], vertices[5], vertices[6], vertices[7]],  # 上面 EFGH
                [vertices[0], vertices[1], vertices[5], vertices[4]],  # 前面 ABFE
                [vertices[2], vertices[3], vertices[7], vertices[6]],  # 後面 CDHG
                [vertices[0], vertices[3], vertices[7], vertices[4]],  # 左面 ADHE
                [vertices[1], vertices[2], vertices[6], vertices[5]]   # 右面 BCGF
            ]
            
            # 面を描画（透明度を設定して内部が見えるようにする）
            for face in faces:
                poly = [[face[j][k] for k in range(3)] for j in range(4)]
                ax.add_collection3d(Poly3DCollection([poly], alpha=0.1, facecolor='lightblue', edgecolor='blue', linewidth=1.5))
            
            # 辺を描画
            edges = [
                [0, 1], [1, 2], [2, 3], [3, 0],  # 底面
                [4, 5], [5, 6], [6, 7], [7, 4],  # 上面
                [0, 4], [1, 5], [2, 6], [3, 7]   # 縦の辺
            ]
            
            for edge in edges:
                points = vertices[edge]
                ax.plot3D(*points.T, 'b-', linewidth=2)
            
            # 頂点ラベル（A,B,C,D,E,F,G,H）
            labels_text = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']
            if labels:
                for i, (vertex, label) in enumerate(zip(vertices, labels_text)):
                    ax.text(vertex[0], vertex[1], vertex[2], label, size=14, color='red', weight='bold')
            
            # 座標軸の設定
            max_range = max(width, depth, height)
            ax.set_xlim([-1, max_range + 1])
            ax.set_ylim([-1, max_range + 1])
            ax.set_zlim([-1, max_range + 1])
            
            # 軸ラベル
            ax.set_xlabel('X')
            ax.set_ylabel('Y')
            ax.set_zlabel('Z')
            
            # 視点を設定
            ax.view_init(elev=20, azim=-75)
            
            # グリッドを表示
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
                shape_type="cuboid"
            )
            
        except Exception as e:
            plt.close()
            print(f"Error drawing 3D cuboid: {e}")
            return GeometryResponse(
                success=False,
                image_base64="",
                shape_type="cuboid"
            )
    
    async def execute_python_code(self, python_code: str) -> dict:
        """Pythonコードを実行して結果を返す（数値計算専用）"""
        print(f"🔍 execute_python_code called")
        print(f"🔍 python_code length: {len(python_code)}")
        
        try:
            # stdoutをキャプチャするための設定
            import sys
            from io import StringIO
            
            # 標準出力をキャプチャ
            old_stdout = sys.stdout
            sys.stdout = captured_output = StringIO()
            
            # 安全な実行環境を準備
            safe_globals = {
                'np': np,
                'numpy': np,
                'math': __import__('math'),
                '__builtins__': {
                    '__import__': __import__,
                    'len': len,
                    'range': range,
                    'enumerate': enumerate,
                    'zip': zip,
                    'map': map,
                    'filter': filter,
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
            
            print(f"🔍 About to execute Python code for calculation")
            print(f"🔍 Python code preview: {python_code[:200]}...")
            
            # Pythonコードを実行
            exec(python_code, safe_globals)
            
            # 標準出力を復元
            sys.stdout = old_stdout
            
            # キャプチャした出力を取得
            output = captured_output.getvalue()
            print(f"✅ Python code executed successfully")
            print(f"📊 Output length: {len(output)}")
            print(f"📊 Output preview: {output[:500]}...")
            
            return {
                "success": True,
                "output": output,
                "error": None
            }
            
        except Exception as e:
            # 標準出力を復元（エラーの場合も必ず復元）
            sys.stdout = old_stdout
            
            error_msg = str(e)
            print(f"❌ Error in execute_python_code: {error_msg}")
            print(f"❌ Error type: {type(e).__name__}")
            
            import traceback
            traceback_str = traceback.format_exc()
            print(f"❌ Traceback: {traceback_str}")
            
            return {
                "success": False,
                "output": "",
                "error": f"{type(e).__name__}: {error_msg}"
            }
