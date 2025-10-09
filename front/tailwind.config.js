/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        'mongene-green': '#8DDB39',
        'mongene-blue': '#11B3F2',
        'mongene-yellow': '#FFCD0A',
        'mongene-ink': '#222222',
        'mongene-muted': '#6b7280',
        'mongene-border': '#e5e7eb',
        'mongene-bg': '#ffffff',
      },
      fontFamily: {
        sans: ['system-ui', '-apple-system', 'Segoe UI', 'Roboto', 'Noto Sans JP', 'Helvetica', 'Arial', 'Apple Color Emoji', 'Segoe UI Emoji', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
