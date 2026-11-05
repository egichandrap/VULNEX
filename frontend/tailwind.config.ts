import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        vulnex: {
          bg: "#0B0F17",
          card: "#121824",
          border: "#1E293B",
          primary: "#3B82F6",
          danger: "#EF4444",
          warning: "#F59E0B",
          success: "#10B981",
        },
      },
    },
  },
  plugins: [],
};

export default config;
