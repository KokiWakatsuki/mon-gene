#!/usr/bin/env python3
"""
図形描画とPDF生成システムのテストスクリプト
"""

import requests
import json
import base64
import os
from typing import Dict, Any

# APIエンドポイント設定
CORE_API_URL = "http://localhost:1234"
BACKEND_API_URL = "http://localhost:8080"

def test_core_api_health():
    """Core APIのヘルスチェック"""
    try:
        response = requests.get(f"{CORE_API_URL}/")
        print(f"✅ Core API Health Check: {response.status_code}")
        print(f"   Response: {response.json()}")
        return True
    except Exception as e:
        print(f"❌ Core API Health Check Failed: {e}")
        return False

def test_backend_api_health():
    """Backend APIのヘルスチェック"""
    try:
        response = requests.get(f"{BACKEND_API_URL}/")
        print(f"✅ Backend API Health Check: {response.status_code}")
        print(f"   Response: {response.json()}")
        return True
    except Exception as e:
        print(f"❌ Backend API Health Check Failed: {e}")
        return False

def test_problem_analysis():
    """問題分析機能のテスト"""
    try:
        test_data = {
            "problem_text": "円柱の体積を求めなさい。底面の半径が3cm、高さが8cmの円柱があります。",
            "unit_parameters": {"geometry": True, "shapes": ["cylinder"]},
            "subject": "math"
        }
        
        response = requests.post(f"{CORE_API_URL}/analyze-problem", json=test_data)
        print(f"✅ Problem Analysis: {response.status_code}")
        result = response.json()
        print(f"   Needs Geometry: {result.get('needs_geometry')}")
        print(f"   Detected Shapes: {result.get('detected_shapes')}")
        return result
    except Exception as e:
        print(f"❌ Problem Analysis Failed: {e}")
        return None

def test_geometry_drawing():
    """図形描画機能のテスト"""
    try:
        # 円柱の描画テスト
        test_data = {
            "shape_type": "cylinder",
            "parameters": {
                "radius": 3,
                "height": 8
            }
        }
        
        response = requests.post(f"{CORE_API_URL}/draw-geometry", json=test_data)
        print(f"✅ Geometry Drawing (Cylinder): {response.status_code}")
        result = response.json()
        
        if result.get('success') and result.get('image_base64'):
            print(f"   Image generated successfully (length: {len(result['image_base64'])})")
            return result['image_base64']
        else:
            print(f"   Failed to generate image")
            return None
            
    except Exception as e:
        print(f"❌ Geometry Drawing Failed: {e}")
        return None

def test_pdf_generation(problem_text: str, image_base64: str = None):
    """PDF生成機能のテスト"""
    try:
        test_data = {
            "problem_text": problem_text,
            "image_base64": image_base64
        }
        
        response = requests.post(f"{CORE_API_URL}/generate-pdf", json=test_data)
        print(f"✅ PDF Generation: {response.status_code}")
        result = response.json()
        
        if result.get('success') and result.get('pdf_base64'):
            print(f"   PDF generated successfully (length: {len(result['pdf_base64'])})")
            
            # PDFファイルを保存
            pdf_data = base64.b64decode(result['pdf_base64'])
            with open('test_output.pdf', 'wb') as f:
                f.write(pdf_data)
            print(f"   PDF saved as 'test_output.pdf'")
            return True
        else:
            print(f"   Failed to generate PDF")
            return False
            
    except Exception as e:
        print(f"❌ PDF Generation Failed: {e}")
        return False

def test_integrated_problem_generation():
    """統合された問題生成機能のテスト"""
    try:
        test_data = {
            "prompt": "円柱の体積を求める問題を作成してください。底面の半径と高さを具体的な数値で指定してください。",
            "subject": "math",
            "filters": {
                "geometry": True,
                "shapes": ["cylinder"],
                "difficulty": 3
            }
        }
        
        response = requests.post(f"{BACKEND_API_URL}/api/generate-problem", json=test_data)
        print(f"✅ Integrated Problem Generation: {response.status_code}")
        result = response.json()
        
        if result.get('success'):
            print(f"   Problem generated successfully")
            print(f"   Content length: {len(result.get('content', ''))}")
            if result.get('image_base64'):
                print(f"   Image included (length: {len(result['image_base64'])})")
            return result
        else:
            print(f"   Failed: {result.get('error')}")
            return None
            
    except Exception as e:
        print(f"❌ Integrated Problem Generation Failed: {e}")
        return None

def test_integrated_pdf_generation():
    """統合されたPDF生成機能のテスト"""
    try:
        test_data = {
            "problem_text": "円柱の体積を求めなさい。\n\n底面の半径が5cm、高さが12cmの円柱があります。\n円柱の体積の公式 V = πr²h を使って計算してください。",
            "image_base64": None  # 実際のテストでは図形描画の結果を使用
        }
        
        response = requests.post(f"{BACKEND_API_URL}/api/generate-pdf", json=test_data)
        print(f"✅ Integrated PDF Generation: {response.status_code}")
        result = response.json()
        
        if result.get('success') and result.get('pdf_base64'):
            print(f"   PDF generated successfully")
            
            # PDFファイルを保存
            pdf_data = base64.b64decode(result['pdf_base64'])
            with open('integrated_test_output.pdf', 'wb') as f:
                f.write(pdf_data)
            print(f"   PDF saved as 'integrated_test_output.pdf'")
            return True
        else:
            print(f"   Failed to generate PDF")
            return False
            
    except Exception as e:
        print(f"❌ Integrated PDF Generation Failed: {e}")
        return False

def main():
    """メインテスト関数"""
    print("🚀 図形描画とPDF生成システムのテストを開始します\n")
    
    # ヘルスチェック
    print("=== ヘルスチェック ===")
    core_healthy = test_core_api_health()
    backend_healthy = test_backend_api_health()
    print()
    
    if not core_healthy:
        print("❌ Core APIが起動していません。先にCore APIを起動してください。")
        print("   実行コマンド: cd core && python main.py")
        return
    
    if not backend_healthy:
        print("❌ Backend APIが起動していません。先にBackend APIを起動してください。")
        print("   実行コマンド: cd back && go run main.go")
        return
    
    # Core API機能テスト
    print("=== Core API機能テスト ===")
    analysis_result = test_problem_analysis()
    image_base64 = test_geometry_drawing()
    pdf_success = test_pdf_generation(
        "円柱の体積を求めなさい。底面の半径が3cm、高さが8cmの円柱があります。",
        image_base64
    )
    print()
    
    # 統合テスト
    print("=== 統合テスト ===")
    problem_result = test_integrated_problem_generation()
    integrated_pdf_success = test_integrated_pdf_generation()
    print()
    
    # 結果サマリー
    print("=== テスト結果サマリー ===")
    tests = [
        ("Core API Health", core_healthy),
        ("Backend API Health", backend_healthy),
        ("Problem Analysis", analysis_result is not None),
        ("Geometry Drawing", image_base64 is not None),
        ("PDF Generation", pdf_success),
        ("Integrated Problem Generation", problem_result is not None),
        ("Integrated PDF Generation", integrated_pdf_success),
    ]
    
    passed = sum(1 for _, result in tests if result)
    total = len(tests)
    
    for test_name, result in tests:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"   {status}: {test_name}")
    
    print(f"\n🎯 テスト結果: {passed}/{total} 通過")
    
    if passed == total:
        print("🎉 すべてのテストが成功しました！")
        print("\n📋 使用したライブラリ:")
        print("   - matplotlib: 図形描画")
        print("   - numpy: 数値計算")
        print("   - reportlab: PDF生成")
        print("   - Pillow: 画像処理")
        print("   - fastapi: Core API")
        print("   - uvicorn: ASGI サーバー")
    else:
        print("⚠️  一部のテストが失敗しました。ログを確認してください。")

if __name__ == "__main__":
    main()
