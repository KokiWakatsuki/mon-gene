from reportlab.lib.pagesizes import A4
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, Image
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.lib.enums import TA_CENTER, TA_LEFT
import base64
import io
from PIL import Image as PILImage

from app.models.pdf import PDFGenerateResponse

class PDFService:
    """PDF生成サービス"""
    
    def __init__(self):
        # 日本語フォントの設定（システムにインストールされている場合）
        try:
            # Noto Sans CJK JP フォントを試す
            pdfmetrics.registerFont(TTFont('NotoSansCJK', '/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc'))
            self.japanese_font = 'NotoSansCJK'
        except:
            try:
                # DejaVu Sans フォントを試す（基本的な日本語サポート）
                pdfmetrics.registerFont(TTFont('DejaVuSans', '/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf'))
                self.japanese_font = 'DejaVuSans'
            except:
                # フォールバック: Helvetica（日本語は表示されない可能性）
                self.japanese_font = 'Helvetica'
    
    async def generate_pdf(self, problem_text: str, image_base64: str = None) -> PDFGenerateResponse:
        """PDFを生成する"""
        try:
            # PDFバッファを作成
            buffer = io.BytesIO()
            
            # PDFドキュメントを作成
            doc = SimpleDocTemplate(buffer, pagesize=A4, 
                                  rightMargin=72, leftMargin=72,
                                  topMargin=72, bottomMargin=18)
            
            # スタイルを設定
            styles = getSampleStyleSheet()
            
            # 日本語対応のスタイルを作成
            title_style = ParagraphStyle(
                'CustomTitle',
                parent=styles['Heading1'],
                fontName=self.japanese_font,
                fontSize=16,
                spaceAfter=30,
                alignment=TA_CENTER
            )
            
            body_style = ParagraphStyle(
                'CustomBody',
                parent=styles['Normal'],
                fontName=self.japanese_font,
                fontSize=12,
                spaceAfter=12,
                alignment=TA_LEFT,
                leading=18
            )
            
            # コンテンツを構築
            story = []
            
            # タイトル
            story.append(Paragraph("数学問題", title_style))
            story.append(Spacer(1, 20))
            
            # 問題文を段落に分割して追加
            paragraphs = problem_text.split('\n\n')
            for para in paragraphs:
                if para.strip():
                    # 改行を<br/>タグに変換
                    formatted_para = para.replace('\n', '<br/>')
                    story.append(Paragraph(formatted_para, body_style))
                    story.append(Spacer(1, 12))
            
            # 図形画像がある場合は追加
            if image_base64 and image_base64 != "mock_image_data":
                try:
                    # Base64画像をデコード
                    image_data = base64.b64decode(image_base64)
                    image_buffer = io.BytesIO(image_data)
                    
                    # PIL Imageで画像を開く
                    pil_image = PILImage.open(image_buffer)
                    
                    # 画像サイズを調整（A4ページに収まるように）
                    max_width = 400
                    max_height = 300
                    
                    # アスペクト比を保持してリサイズ
                    pil_image.thumbnail((max_width, max_height), PILImage.Resampling.LANCZOS)
                    
                    # ReportLab用の画像バッファを作成
                    img_buffer = io.BytesIO()
                    pil_image.save(img_buffer, format='PNG')
                    img_buffer.seek(0)
                    
                    # ReportLab Imageオブジェクトを作成
                    img = Image(img_buffer, width=pil_image.width, height=pil_image.height)
                    
                    story.append(Spacer(1, 20))
                    story.append(img)
                    story.append(Spacer(1, 20))
                    
                except Exception as e:
                    # 画像処理エラーの場合はスキップ
                    print(f"Image processing error: {e}")
            
            # PDFを生成
            doc.build(story)
            
            # Base64エンコード
            buffer.seek(0)
            pdf_base64 = base64.b64encode(buffer.getvalue()).decode()
            
            return PDFGenerateResponse(
                success=True,
                pdf_base64=pdf_base64
            )
            
        except Exception as e:
            print(f"PDF generation error: {e}")
            return PDFGenerateResponse(
                success=False,
                pdf_base64=""
            )
