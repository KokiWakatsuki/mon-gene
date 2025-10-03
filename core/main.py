from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, Any, Optional
import matplotlib
matplotlib.use('Agg')  # GUI不要のバックエンドを使用
import matplotlib.pyplot as plt
import matplotlib.patches as patches
import numpy as np
import io
import base64
from reportlab.lib.pagesizes import A4
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, Image
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
import os
import tempfile
import json

app = FastAPI()

class ProblemRequest(BaseModel):
    problem_text: str
    unit_parameters: Dict[str, Any]
    subject: str = "math"

class GeometryDrawRequest(BaseModel):
    shape_type: str
    parameters: Dict[str, Any]
    labels: Optional[Dict[str, str]] = None

class CustomGeometryRequest(BaseModel):
    python_code: str
    problem_text: str

class PDFGenerateRequest(BaseModel):
    problem_text: str
    image_base64: Optional[str] = None

# 日本語フォントの設定（システムにインストールされている場合）
def setup_japanese_font():
    try:
        # macOSの場合
        font_paths = [
            "/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
            "/System/Library/Fonts/Hiragino Sans GB.ttc",
            "/Library/Fonts/Arial Unicode MS.ttf"
        ]
        
        for font_path in font_paths:
            if os.path.exists(font_path):
                pdfmetrics.registerFont(TTFont('Japanese', font_path))
                return True
        
        # フォントが見つからない場合はデフォルトを使用
        return False
    except:
        return False

# 図形描画関数
def draw_prism(base_area: float, height: float, shape_type: str = "rectangular"):
    """角柱を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    if shape_type == "rectangular":
        # 直方体の描画
        # 底面
        bottom = patches.Rectangle((1, 1), 3, 2, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
        ax.add_patch(bottom)
        
        # 上面
        top = patches.Rectangle((1.5, 2.5), 3, 2, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
        ax.add_patch(top)
        
        # 側面
        ax.plot([1, 1.5], [1, 2.5], 'k-', linewidth=2)
        ax.plot([4, 4.5], [1, 2.5], 'k-', linewidth=2)
        ax.plot([4, 4.5], [3, 4.5], 'k-', linewidth=2)
        ax.plot([1, 1.5], [3, 4.5], 'k-', linewidth=2)
        
        # 破線（隠れた線）
        ax.plot([1.5, 1.5], [2.5, 4.5], 'k--', linewidth=1, alpha=0.7)
        ax.plot([1.5, 4.5], [2.5, 2.5], 'k--', linewidth=1, alpha=0.7)
        
        # ラベル
        ax.text(2.5, 0.5, f'底面積 S', fontsize=12, ha='center')
        ax.text(5, 3.5, f'高さ h', fontsize=12, ha='center')
        
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax.set_aspect('equal')
    ax.axis('off')
    ax.set_title('角柱', fontsize=14, fontweight='bold')
    
    return fig

def draw_cylinder(radius: float, height: float):
    """円柱を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 底面の楕円
    bottom_ellipse = patches.Ellipse((3, 1.5), 3, 1, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(bottom_ellipse)
    
    # 上面の楕円
    top_ellipse = patches.Ellipse((3, 4), 3, 1, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(top_ellipse)
    
    # 側面
    ax.plot([1.5, 1.5], [1.5, 4], 'k-', linewidth=2)
    ax.plot([4.5, 4.5], [1.5, 4], 'k-', linewidth=2)
    
    # 破線（隠れた線）
    theta = np.linspace(np.pi, 2*np.pi, 100)
    x_hidden = 3 + 1.5 * np.cos(theta)
    y_hidden = 1.5 + 0.5 * np.sin(theta)
    ax.plot(x_hidden, y_hidden, 'k--', linewidth=1, alpha=0.7)
    
    # 中心点
    ax.plot(3, 1.5, 'ko', markersize=3)
    
    # ラベル
    ax.text(3, 0.8, f'底面積 S', fontsize=12, ha='center')
    ax.text(5.2, 2.75, f'高さ h', fontsize=12, ha='center')
    ax.text(3.8, 1.3, f'半径 r', fontsize=12, ha='center')
    
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax.set_aspect('equal')
    ax.axis('off')
    ax.set_title('円柱', fontsize=14, fontweight='bold')
    
    return fig

def draw_triangle(vertices: list, labels: dict = None):
    """三角形を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 三角形の描画
    triangle = patches.Polygon(vertices, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(triangle)
    
    # 頂点のラベル
    if labels:
        for i, (x, y) in enumerate(vertices):
            label = labels.get(f'vertex_{i}', f'P{i}')
            ax.text(x, y, label, fontsize=12, ha='center', va='center', 
                   bbox=dict(boxstyle="circle,pad=0.1", facecolor='white', edgecolor='black'))
    
    ax.set_xlim(-1, 6)
    ax.set_ylim(-1, 5)
    ax.set_aspect('equal')
    ax.grid(True, alpha=0.3)
    ax.set_title('三角形', fontsize=14, fontweight='bold')
    
    return fig

def draw_circle(radius: float, center: tuple = (0, 0)):
    """円を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    circle = patches.Circle(center, radius, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(circle)
    
    # 中心点
    ax.plot(center[0], center[1], 'ko', markersize=5)
    
    # 半径の線
    ax.plot([center[0], center[0] + radius], [center[1], center[1]], 'k-', linewidth=2)
    ax.text(center[0] + radius/2, center[1] + 0.2, f'r = {radius}', fontsize=12, ha='center')
    
    ax.set_xlim(center[0] - radius - 1, center[0] + radius + 1)
    ax.set_ylim(center[1] - radius - 1, center[1] + radius + 1)
    ax.set_aspect('equal')
    ax.grid(True, alpha=0.3)
    ax.set_title('円', fontsize=14, fontweight='bold')
    
    return fig

def draw_square(side_length: float = 8, show_moving_points: bool = True):
    """正方形を描画（移動する点付き）"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 正方形の描画
    square = patches.Rectangle((0, 0), side_length, side_length, 
                              linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(square)
    
    # 頂点のラベル
    ax.text(-0.3, -0.3, 'A', fontsize=14, ha='center', va='center', fontweight='bold')
    ax.text(side_length + 0.3, -0.3, 'B', fontsize=14, ha='center', va='center', fontweight='bold')
    ax.text(side_length + 0.3, side_length + 0.3, 'C', fontsize=14, ha='center', va='center', fontweight='bold')
    ax.text(-0.3, side_length + 0.3, 'D', fontsize=14, ha='center', va='center', fontweight='bold')
    
    # 移動する点を表示
    if show_moving_points:
        # 点P（辺BC上）
        p_x = side_length * 0.6  # 例として60%の位置
        ax.plot(side_length, p_x, 'ro', markersize=8)
        ax.text(side_length + 0.5, p_x, 'P', fontsize=12, ha='left', va='center', fontweight='bold', color='red')
        
        # 点Q（辺CD上）
        q_y = side_length * 0.4  # 例として40%の位置
        ax.plot(side_length - q_y, side_length, 'go', markersize=8)
        ax.text(side_length - q_y, side_length + 0.5, 'Q', fontsize=12, ha='center', va='bottom', fontweight='bold', color='green')
        
        # PQを結ぶ線
        ax.plot([side_length, side_length - q_y], [p_x, side_length], 'r--', linewidth=2, alpha=0.7)
        
        # 移動方向の矢印
        ax.annotate('', xy=(side_length, p_x + 1), xytext=(side_length, p_x - 1),
                   arrowprops=dict(arrowstyle='->', color='red', lw=2))
        ax.annotate('', xy=(side_length - q_y - 1, side_length), xytext=(side_length - q_y + 1, side_length),
                   arrowprops=dict(arrowstyle='->', color='green', lw=2))
    
    # 辺の長さ表示
    ax.text(side_length/2, -0.8, f'{side_length}cm', fontsize=12, ha='center', va='center')
    
    ax.set_xlim(-1, side_length + 2)
    ax.set_ylim(-1, side_length + 2)
    ax.set_aspect('equal')
    ax.grid(True, alpha=0.3)
    ax.set_title('Square ABCD', fontsize=14, fontweight='bold')
    
    return fig

def draw_cube(side_length: float = 4):
    """立方体を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 立方体の頂点座標（3D風の2D表現）
    # 前面の正方形
    front_square = patches.Rectangle((1, 1), side_length, side_length, 
                                   linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(front_square)
    
    # 背面の正方形（オフセット）
    offset = side_length * 0.3
    back_square = patches.Rectangle((1 + offset, 1 + offset), side_length, side_length, 
                                  linewidth=2, edgecolor='black', facecolor='lightcyan', alpha=0.5)
    ax.add_patch(back_square)
    
    # 接続線（エッジ）
    # 前面から背面への線
    ax.plot([1, 1 + offset], [1, 1 + offset], 'k-', linewidth=2)  # A to E
    ax.plot([1 + side_length, 1 + side_length + offset], [1, 1 + offset], 'k-', linewidth=2)  # B to F
    ax.plot([1 + side_length, 1 + side_length + offset], [1 + side_length, 1 + side_length + offset], 'k-', linewidth=2)  # C to G
    ax.plot([1, 1 + offset], [1 + side_length, 1 + side_length + offset], 'k-', linewidth=2)  # D to H
    
    # 隠れた線（破線）
    ax.plot([1 + offset, 1 + offset], [1 + offset, 1 + side_length + offset], 'k--', linewidth=1, alpha=0.7)
    ax.plot([1 + offset, 1 + side_length + offset], [1 + offset, 1 + offset], 'k--', linewidth=1, alpha=0.7)
    
    # 頂点のラベル
    ax.text(1 - 0.2, 1 - 0.2, 'A', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 + side_length + 0.2, 1 - 0.2, 'B', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 + side_length + 0.2, 1 + side_length + 0.2, 'C', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 - 0.2, 1 + side_length + 0.2, 'D', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 + offset + 0.2, 1 + offset + 0.2, 'E', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 + side_length + offset + 0.2, 1 + offset - 0.2, 'F', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 + side_length + offset + 0.2, 1 + side_length + offset + 0.2, 'G', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(1 + offset - 0.2, 1 + side_length + offset + 0.2, 'H', fontsize=12, ha='center', va='center', fontweight='bold')
    
    ax.set_xlim(0, 2 + side_length + offset)
    ax.set_ylim(0, 2 + side_length + offset)
    ax.set_aspect('equal')
    ax.axis('off')
    ax.set_title('Cube', fontsize=14, fontweight='bold')
    
    return fig

def draw_sphere(radius: float = 2):
    """球を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 球の外形（円）
    sphere = patches.Circle((3, 3), radius, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(sphere)
    
    # 中心点
    ax.plot(3, 3, 'ko', markersize=5)
    
    # 半径の線
    ax.plot([3, 3 + radius], [3, 3], 'k-', linewidth=2)
    ax.text(3 + radius/2, 3 + 0.3, f'r = {radius}', fontsize=12, ha='center')
    
    # 球の立体感を出すための楕円（緯度線）
    ellipse1 = patches.Ellipse((3, 3), radius * 2, radius * 0.6, linewidth=1, edgecolor='gray', facecolor='none', alpha=0.5)
    ax.add_patch(ellipse1)
    ellipse2 = patches.Ellipse((3, 3), radius * 1.4, radius * 0.4, linewidth=1, edgecolor='gray', facecolor='none', alpha=0.5)
    ax.add_patch(ellipse2)
    
    # 経度線（縦の楕円）
    ellipse3 = patches.Ellipse((3, 3), radius * 0.6, radius * 2, linewidth=1, edgecolor='gray', facecolor='none', alpha=0.5)
    ax.add_patch(ellipse3)
    
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 6)
    ax.set_aspect('equal')
    ax.grid(True, alpha=0.3)
    ax.set_title('Sphere', fontsize=14, fontweight='bold')
    
    return fig

def draw_cone(radius: float = 2, height: float = 4):
    """円錐を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 底面の楕円
    base_ellipse = patches.Ellipse((3, 1), radius * 2, radius * 0.6, linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(base_ellipse)
    
    # 円錐の側面（三角形）
    apex_x, apex_y = 3, 1 + height
    left_base = 3 - radius
    right_base = 3 + radius
    
    # 左側の線
    ax.plot([left_base, apex_x], [1, apex_y], 'k-', linewidth=2)
    # 右側の線
    ax.plot([right_base, apex_x], [1, apex_y], 'k-', linewidth=2)
    
    # 隠れた底面の線（破線）
    theta = np.linspace(np.pi, 2*np.pi, 100)
    x_hidden = 3 + radius * np.cos(theta)
    y_hidden = 1 + radius * 0.3 * np.sin(theta)
    ax.plot(x_hidden, y_hidden, 'k--', linewidth=1, alpha=0.7)
    
    # 中心点と頂点
    ax.plot(3, 1, 'ko', markersize=3)
    ax.plot(apex_x, apex_y, 'ko', markersize=3)
    
    # ラベル
    ax.text(3, 0.5, f'r = {radius}', fontsize=12, ha='center')
    ax.text(apex_x + 0.3, apex_y, 'A', fontsize=12, ha='center', fontweight='bold')
    ax.text(3 + 0.3, 1, 'O', fontsize=12, ha='center', fontweight='bold')
    
    # 高さの線
    ax.plot([3, apex_x], [1, apex_y], 'r--', linewidth=1, alpha=0.7)
    ax.text(3.3, (1 + apex_y)/2, f'h = {height}', fontsize=12, ha='left', color='red')
    
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 6)
    ax.set_aspect('equal')
    ax.axis('off')
    ax.set_title('Cone', fontsize=14, fontweight='bold')
    
    return fig

def draw_pyramid(base_side: float = 3, height: float = 4):
    """四角錐を描画"""
    fig, ax = plt.subplots(1, 1, figsize=(8, 6))
    
    # 底面の正方形（3D風）
    base_square = patches.Polygon([(2, 1), (2 + base_side, 1), (2 + base_side + 0.5, 1.5), (2 + 0.5, 1.5)], 
                                linewidth=2, edgecolor='black', facecolor='lightblue', alpha=0.7)
    ax.add_patch(base_square)
    
    # 頂点の座標
    apex_x, apex_y = 2 + base_side/2 + 0.25, 1 + height
    
    # 四角錐の辺
    # 前面の辺
    ax.plot([2, apex_x], [1, apex_y], 'k-', linewidth=2)  # A to S
    ax.plot([2 + base_side, apex_x], [1, apex_y], 'k-', linewidth=2)  # B to S
    
    # 背面の辺
    ax.plot([2 + base_side + 0.5, apex_x], [1.5, apex_y], 'k-', linewidth=2)  # C to S
    ax.plot([2 + 0.5, apex_x], [1.5, apex_y], 'k-', linewidth=2)  # D to S
    
    # 隠れた底面の辺（破線）
    ax.plot([2, 2 + 0.5], [1, 1.5], 'k--', linewidth=1, alpha=0.7)
    ax.plot([2 + base_side, 2 + base_side + 0.5], [1, 1.5], 'k--', linewidth=1, alpha=0.7)
    
    # 頂点のラベル
    ax.text(2 - 0.2, 1 - 0.2, 'A', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(2 + base_side + 0.2, 1 - 0.2, 'B', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(2 + base_side + 0.5 + 0.2, 1.5 + 0.2, 'C', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(2 + 0.5 - 0.2, 1.5 + 0.2, 'D', fontsize=12, ha='center', va='center', fontweight='bold')
    ax.text(apex_x + 0.2, apex_y + 0.2, 'S', fontsize=12, ha='center', va='center', fontweight='bold')
    
    # 高さの線
    base_center_x = 2 + base_side/2 + 0.25
    base_center_y = 1.25
    ax.plot([base_center_x, apex_x], [base_center_y, apex_y], 'r--', linewidth=1, alpha=0.7)
    ax.text(base_center_x + 0.3, (base_center_y + apex_y)/2, f'h = {height}', fontsize=12, ha='left', color='red')
    
    ax.set_xlim(1, 7)
    ax.set_ylim(0, 6)
    ax.set_aspect('equal')
    ax.axis('off')
    ax.set_title('Square Pyramid', fontsize=14, fontweight='bold')
    
    return fig

def figure_to_base64(fig):
    """matplotlib図をbase64文字列に変換"""
    buffer = io.BytesIO()
    fig.savefig(buffer, format='png', dpi=150, bbox_inches='tight')
    buffer.seek(0)
    image_base64 = base64.b64encode(buffer.getvalue()).decode()
    plt.close(fig)
    return image_base64

def execute_custom_geometry_code(python_code: str) -> plt.Figure:
    """カスタムPythonコードを安全に実行して図形を生成"""
    # フォント設定（利用可能なフォントのみ使用）
    plt.rcParams['font.family'] = ['DejaVu Sans', 'sans-serif']
    
    # import文を除去してコードをクリーンアップ
    lines = python_code.split('\n')
    cleaned_lines = []
    
    for line in lines:
        stripped_line = line.strip()
        # import文をスキップ
        if (stripped_line.startswith('import ') or 
            stripped_line.startswith('from ') or
            stripped_line == ''):
            continue
        cleaned_lines.append(line)
    
    cleaned_code = '\n'.join(cleaned_lines)
    
    # 安全な実行環境を準備
    import builtins
    safe_builtins = {
        'len': builtins.len,
        'range': builtins.range,
        'enumerate': builtins.enumerate,
        'zip': builtins.zip,
        'list': builtins.list,
        'tuple': builtins.tuple,
        'dict': builtins.dict,
        'set': builtins.set,
        'str': builtins.str,
        'int': builtins.int,
        'float': builtins.float,
        'bool': builtins.bool,
        'min': builtins.min,
        'max': builtins.max,
        'abs': builtins.abs,
        'round': builtins.round,
        'sum': builtins.sum,
        'print': builtins.print,
    }
    
    # 3D描画用のモジュールをインポート
    from mpl_toolkits.mplot3d import Axes3D
    from mpl_toolkits.mplot3d.art3d import Poly3DCollection
    import mpl_toolkits.mplot3d as mplot3d
    
    safe_globals = {
        '__builtins__': safe_builtins,
        'matplotlib': matplotlib,
        'plt': plt,
        'patches': patches,
        'np': np,
        'numpy': np,
        # 3D描画用モジュール
        'Axes3D': Axes3D,
        'Poly3DCollection': Poly3DCollection,
        'mplot3d': mplot3d,
        # 数学関数も追加
        'sin': np.sin,
        'cos': np.cos,
        'tan': np.tan,
        'pi': np.pi,
        'sqrt': np.sqrt,
        'array': np.array,
        'linspace': np.linspace,
        'arange': np.arange,
        'meshgrid': np.meshgrid,
        'isclose': np.isclose,
        'arctan2': np.arctan2,
        'argsort': np.argsort,
        'argmax': np.argmax,
        'zeros_like': np.zeros_like,
        'copy': np.copy,
        'nan': np.nan,
    }
    
    safe_locals = {}
    
    try:
        # クリーンアップされたコードを実行
        exec(cleaned_code, safe_globals, safe_locals)
        
        # figオブジェクトを取得
        if 'fig' in safe_locals:
            return safe_locals['fig']
        else:
            # figが見つからない場合は現在のfigureを取得
            return plt.gcf()
            
    except Exception as e:
        # エラーが発生した場合はエラーメッセージを含む図を作成
        fig, ax = plt.subplots(1, 1, figsize=(8, 6))
        ax.text(0.5, 0.5, f'Code execution error:\n{str(e)}', 
                ha='center', va='center', fontsize=12, 
                bbox=dict(boxstyle="round,pad=0.3", facecolor='lightcoral'))
        ax.set_xlim(0, 1)
        ax.set_ylim(0, 1)
        ax.axis('off')
        ax.set_title('Geometry Generation Error', fontsize=14, fontweight='bold')
        return fig

@app.post("/draw-geometry")
async def draw_geometry(request: GeometryDrawRequest):
    """図形描画エンドポイント"""
    try:
        shape_type = request.shape_type.lower()
        params = request.parameters
        
        if shape_type == "prism" or shape_type == "角柱":
            fig = draw_prism(
                base_area=params.get("base_area", 10),
                height=params.get("height", 5),
                shape_type=params.get("prism_type", "rectangular")
            )
        elif shape_type == "cylinder" or shape_type == "円柱":
            fig = draw_cylinder(
                radius=params.get("radius", 2),
                height=params.get("height", 4)
            )
        elif shape_type == "triangle" or shape_type == "三角形":
            vertices = params.get("vertices", [[0, 0], [3, 0], [1.5, 3]])
            fig = draw_triangle(vertices, request.labels)
        elif shape_type == "circle" or shape_type == "円":
            fig = draw_circle(
                radius=params.get("radius", 2),
                center=params.get("center", (0, 0))
            )
        elif shape_type == "square" or shape_type == "正方形":
            fig = draw_square(
                side_length=params.get("side_length", 8),
                show_moving_points=params.get("show_moving_points", True)
            )
        elif shape_type == "cube" or shape_type == "立方体":
            fig = draw_cube(
                side_length=params.get("side_length", 4)
            )
        elif shape_type == "sphere" or shape_type == "球":
            fig = draw_sphere(
                radius=params.get("radius", 2)
            )
        elif shape_type == "cone" or shape_type == "円錐":
            fig = draw_cone(
                radius=params.get("radius", 2),
                height=params.get("height", 4)
            )
        elif shape_type == "pyramid" or shape_type == "四角錐":
            fig = draw_pyramid(
                base_side=params.get("base_side", 3),
                height=params.get("height", 4)
            )
        else:
            raise HTTPException(status_code=400, detail=f"Unsupported shape type: {shape_type}")
        
        image_base64 = figure_to_base64(fig)
        
        return {
            "success": True,
            "image_base64": image_base64,
            "shape_type": shape_type
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/generate-pdf")
async def generate_pdf(request: PDFGenerateRequest):
    """PDF生成エンドポイント"""
    try:
        # 一時ファイルを作成
        with tempfile.NamedTemporaryFile(delete=False, suffix='.pdf') as tmp_file:
            pdf_path = tmp_file.name
        
        # PDF文書を作成
        doc = SimpleDocTemplate(pdf_path, pagesize=A4)
        story = []
        
        # スタイルの設定
        styles = getSampleStyleSheet()
        
        # 日本語フォントの設定を試行
        japanese_font_available = setup_japanese_font()
        
        if japanese_font_available:
            title_style = ParagraphStyle(
                'CustomTitle',
                parent=styles['Heading1'],
                fontName='Japanese',
                fontSize=16,
                spaceAfter=20,
                alignment=1  # 中央揃え
            )
            normal_style = ParagraphStyle(
                'CustomNormal',
                parent=styles['Normal'],
                fontName='Japanese',
                fontSize=12,
                spaceAfter=12
            )
        else:
            title_style = styles['Heading1']
            normal_style = styles['Normal']
        
        # タイトル
        title = Paragraph("数学問題", title_style)
        story.append(title)
        story.append(Spacer(1, 20))
        
        # 問題文
        problem_paragraph = Paragraph(request.problem_text, normal_style)
        story.append(problem_paragraph)
        story.append(Spacer(1, 20))
        
        # 画像がある場合は追加
        if request.image_base64:
            # base64画像を一時ファイルに保存
            image_data = base64.b64decode(request.image_base64)
            with tempfile.NamedTemporaryFile(delete=False, suffix='.png') as img_tmp:
                img_tmp.write(image_data)
                img_path = img_tmp.name
            
            # 画像をPDFに追加
            img = Image(img_path, width=4*inch, height=3*inch)
            story.append(img)
            
            # 一時画像ファイルを削除
            os.unlink(img_path)
        
        # PDFを構築
        doc.build(story)
        
        # PDFファイルを読み込んでbase64エンコード
        with open(pdf_path, 'rb') as pdf_file:
            pdf_base64 = base64.b64encode(pdf_file.read()).decode()
        
        # 一時PDFファイルを削除
        os.unlink(pdf_path)
        
        return {
            "success": True,
            "pdf_base64": pdf_base64
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/draw-custom-geometry")
async def draw_custom_geometry(request: CustomGeometryRequest):
    """カスタムPythonコードを実行して図形を描画"""
    try:
        fig = execute_custom_geometry_code(request.python_code)
        image_base64 = figure_to_base64(fig)
        
        return {
            "success": True,
            "image_base64": image_base64,
            "problem_text": request.problem_text
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/analyze-problem")
async def analyze_problem(request: ProblemRequest):
    """問題を分析して図形が必要かどうかを判定"""
    try:
        problem_text = request.problem_text.lower()
        unit_params = request.unit_parameters
        
        # 図形関連のキーワードを検出（空間図形を追加）
        geometry_keywords = [
            "角柱", "円柱", "体積", "底面積", "高さ", "半径",
            "三角形", "円", "図形", "面積", "周囲", "直径",
            "正方形", "長方形", "四角形", "多角形", "座標", "点",
            "辺", "頂点", "移動", "直線", "線分", "距離",
            "立方体", "球", "円錐", "四角錐", "空間図形", "立体",
            "prism", "cylinder", "volume", "area", "radius", "height",
            "square", "rectangle", "polygon", "coordinate", "point",
            "side", "vertex", "move", "line", "segment", "distance",
            "cube", "sphere", "cone", "pyramid", "spatial", "solid"
        ]
        
        needs_geometry = False
        detected_shapes = []
        
        # 単元パラメータから図形の必要性を判定
        if ("geometry" in str(unit_params).lower() or 
            "図形" in str(unit_params) or 
            "spatial_geometry" in str(unit_params).lower() or
            "空間図形" in str(unit_params)):
            needs_geometry = True
        
        # 問題文から図形キーワードを検出
        for keyword in geometry_keywords:
            if keyword in problem_text or keyword in str(unit_params).lower():
                needs_geometry = True
                if keyword in ["角柱", "prism"]:
                    detected_shapes.append("prism")
                elif keyword in ["円柱", "cylinder"]:
                    detected_shapes.append("cylinder")
                elif keyword in ["三角形", "triangle"]:
                    detected_shapes.append("triangle")
                elif keyword in ["円", "circle"]:
                    detected_shapes.append("circle")
                elif keyword in ["正方形", "square"]:
                    detected_shapes.append("square")
                elif keyword in ["立方体", "cube"]:
                    detected_shapes.append("cube")
                elif keyword in ["球", "sphere"]:
                    detected_shapes.append("sphere")
                elif keyword in ["円錐", "cone"]:
                    detected_shapes.append("cone")
                elif keyword in ["四角錐", "pyramid"]:
                    detected_shapes.append("pyramid")
        
        # 重複を除去
        detected_shapes = list(set(detected_shapes))
        
        return {
            "success": True,
            "needs_geometry": needs_geometry,
            "detected_shapes": detected_shapes,
            "suggested_parameters": {
                "prism": {"base_area": 10, "height": 5, "prism_type": "rectangular"},
                "cylinder": {"radius": 2, "height": 4},
                "triangle": {"vertices": [[0, 0], [3, 0], [1.5, 3]]},
                "circle": {"radius": 2, "center": [0, 0]},
                "square": {"side_length": 8, "show_moving_points": True},
                "cube": {"side_length": 4},
                "sphere": {"radius": 2},
                "cone": {"radius": 2, "height": 4},
                "pyramid": {"base_side": 3, "height": 4}
            }
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def root():
    return {"message": "Mongene Core API Server is running"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=1234)
