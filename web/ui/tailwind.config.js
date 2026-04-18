/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          50:  '#f0f4ff',
          100: '#dce6ff',
          200: '#b9cdff',
          500: '#4f7bff',
          600: '#3a62e0',
          700: '#2a4bbf',
          900: '#1a2d7a',
        },
      },
    },
  },
  plugins: [],
}
