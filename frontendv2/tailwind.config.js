/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "primary": "#00f5d4", // Vibrant Teal
        "background-light": "#f6f6f8",
        "background-dark": "#1e1e21", // Dark Charcoal
        "panel-dark": "#2a2a2e", // Slightly lighter Gray
        "row-odd-dark": "#2f2f33",
        "text-light": "#e0e0e0", // Off-White
        "text-subtle-dark": "#9da6b9",
        "border-dark": "#3b4354",
        "error": "#ff4d4f",
        "warning": "#ffc107",
        "info": "#4caf50",
        "debug": "#2196f3"
      },
      fontFamily: {
        "display": ["Space Grotesk", "sans-serif"],
        "mono": ["ui-monospace", "SFMono-Regular", "Menlo", "Monaco", "Consolas", "Liberation Mono", "Courier New", "monospace"]
      },
      borderRadius: {
        "DEFAULT": "0.25rem",
        "lg": "0.5rem",
        "xl": "0.75rem",
        "full": "9999px"
      },
    },
  },
  plugins: [],
}
