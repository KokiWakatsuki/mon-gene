import type { Metadata } from "next";
import { Noto_Sans_JP } from "next/font/google";
import "./globals.css";

const notoSansJP = Noto_Sans_JP({
  weight: ['400', '500', '700'],
  subsets: ["latin"],
  variable: "--font-noto-sans-jp",
  display: 'swap',
});

export const metadata: Metadata = {
  title: "モンジェネ",
  description: "数学問題生成アプリケーション",
  icons: {
    icon: '/images/モンジェネロゴ.svg',
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className={`${notoSansJP.variable} antialiased`}>
        {children}
      </body>
    </html>
  );
}
