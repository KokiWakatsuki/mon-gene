'use client';

import React, { useEffect } from 'react';

export default function BackgroundShapes() {
  useEffect(() => {
    const container = document.getElementById('bg-container');
    if (!container) return;

    // SVGシェイプのデータURI
    const shapeUrls = [
      "data:image/svg+xml,%3Csvg viewBox='0 0 243 122' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M237.266,121.4l-231.731,0c-3.057,-0 -5.534,-2.478 -5.534,-5.534l0,-110.331c0,-3.057 2.478,-5.534 5.534,-5.534l231.731,0c3.057,0 5.534,2.478 5.534,5.534l0,110.331c-0,3.057 -2.478,5.534 -5.534,5.534Z' fill='black'/%3E%3C/svg%3E",
      "data:image/svg+xml,%3Csvg viewBox='0 0 243 122' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M64.833,116.376c-0.25,2.838 -2.626,5.014 -5.475,5.014c-14.469,0.01 -53.823,0.01 -53.823,0.01c-3.057,-0 -5.534,-2.478 -5.534,-5.534l0,-110.331c0,-3.057 2.478,-5.534 5.534,-5.534l231.731,0c3.057,0 5.534,2.478 5.534,5.534l0,110.331c-0,3.057 -2.478,5.534 -5.534,5.534c-0,0 -39.356,0 -53.823,0c-2.854,-0 -5.235,-2.18 -5.485,-5.023c-2.537,-28.993 -26.909,-51.764 -56.557,-51.764c-29.649,0 -54.021,22.771 -56.567,51.763Z' fill='black'/%3E%3C/svg%3E",
      "data:image/svg+xml,%3Csvg viewBox='0 0 113 57' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M5.506,56.787c-1.565,-0 -3.056,-0.666 -4.1,-1.831c-1.044,-1.165 -1.544,-2.72 -1.373,-4.275c3.059,-28.474 27.192,-50.681 56.471,-50.681c29.279,0 53.412,22.207 56.462,50.682c0.17,1.553 -0.328,3.105 -1.371,4.268c-1.043,1.163 -2.531,1.828 -4.093,1.828c-19.479,0.009 -82.516,0.009 -101.995,0.009Z' fill='black'/%3E%3C/svg%3E",
      "data:image/svg+xml,%3Csvg viewBox='0 0 122 122' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M115.866,0c3.057,0 5.534,2.478 5.534,5.534l0,110.331c-0,3.057 -2.478,5.534 -5.534,5.534l-110.331,0c-3.057,-0 -5.534,-2.478 -5.534,-5.534l-0,-110.331c0,-3.057 2.478,-5.534 5.534,-5.534l110.331,0Z' fill='black'/%3E%3C/svg%3E",
      "data:image/svg+xml,%3Csvg viewBox='0 0 114 114' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M111.952,104.126c1.583,1.583 2.056,3.963 1.2,6.031c-0.857,2.068 -2.875,3.416 -5.113,3.416l-102.505,-0c-3.057,-0 -5.534,-2.478 -5.534,-5.534l0,-102.505c-0,-2.238 1.348,-4.256 3.416,-5.113c2.068,-0.857 4.448,-0.383 6.031,1.2c22.652,22.652 79.852,79.852 102.505,102.505Z' fill='black'/%3E%3C/svg%3E",
      "data:image/svg+xml,%3Csvg viewBox='0 0 225 118' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M7.968,117.416c-1.707,0.853 -3.734,0.762 -5.357,-0.241c-1.623,-1.003 -2.611,-2.775 -2.611,-4.684l-0,-106.985c0,-3.041 2.465,-5.506 5.506,-5.506c28.489,-0 159.911,-0 213.97,-0c2.554,-0 4.772,1.756 5.359,4.241c0.587,2.485 -0.612,5.048 -2.896,6.19c-51.054,25.527 -180.549,90.274 -213.97,106.985Z' fill='black'/%3E%3C/svg%3E",
      "data:image/svg+xml,%3Csvg viewBox='0 0 114 114' xmlns='http://www.w3.org/2000/svg'%3E%3Ccircle cx='56.787' cy='56.787' r='56.787' fill='black'/%3E%3C/svg%3E",
    ];

    const colorClasses = ['color-green', 'color-yellow', 'color-blue', 'color-pink'];

    // グリッドシステムで配置
    const cols = 6;
    const rows = 5;
    const xStep = 100 / cols;
    const yStep = 100 / rows;

    for (let r = 0; r < rows; r++) {
      for (let c = 0; c < cols; c++) {
        // 50%の確率でスキップ
        if (Math.random() > 0.5) continue;

        const div = document.createElement('div');
        div.classList.add('bg-shape');

        const randomShapeUrl = shapeUrls[Math.floor(Math.random() * shapeUrls.length)];
        div.style.webkitMaskImage = `url("${randomShapeUrl}")`;
        div.style.maskImage = `url("${randomShapeUrl}")`;

        const randomColorClass = colorClasses[Math.floor(Math.random() * colorClasses.length)];
        div.classList.add(randomColorClass);

        const baseX = c * xStep;
        const baseY = r * yStep;
        const jitterX = Math.random() * (xStep * 0.6);
        const jitterY = Math.random() * (yStep * 0.6);
        const sizeBase = 4 + Math.random() * 4;
        const randomRotate = Math.random() * 360;

        div.style.left = `${baseX + jitterX}%`;
        div.style.top = `${baseY + jitterY}%`;
        div.style.width = `${sizeBase}vw`;
        div.style.height = `${sizeBase}vw`;
        div.style.transform = `rotate(${randomRotate}deg)`;

        container.appendChild(div);
      }
    }
  }, []);

  return (
    <>
      <div id="bg-container" className="fixed top-0 left-0 w-screen h-screen overflow-hidden pointer-events-none bg-gray-50" style={{ zIndex: 'var(--z-bg)' }} />
      <div
        className="fixed bg-white/75 backdrop-blur rounded-[20px] shadow-[0_4px_30px_rgba(0,0,0,0.05)] pointer-events-none top-5 left-5 right-5 bottom-5 max-md:top-2.5 max-md:left-2.5 max-md:right-2.5 max-md:bottom-2.5"
        style={{ zIndex: 'var(--z-bg)' }}
      />
    </>
  );
}
