/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#f6831e',
        'background-light': '#f9f7f2',
        'background-dark': '#221810',
        surface: '#ffffff',
        charcoal: '#3e3935',
      },
      fontFamily: {
        display: ['Inter', 'system-ui', 'sans-serif'],
      },
      borderRadius: {
        DEFAULT: '0.25rem',
        lg: '0.5rem',
        xl: '0.75rem',
        full: '9999px',
      },
      boxShadow: {
        sm: '0 1px 2px rgba(62, 57, 53, 0.05)',
        md: '0 4px 6px rgba(62, 57, 53, 0.07)',
        lg: '0 10px 15px rgba(62, 57, 53, 0.1)',
      },
      transitionDuration: {
        fast: '150ms',
        normal: '250ms',
        slow: '350ms',
      }
    },
  },
  plugins: [],
}