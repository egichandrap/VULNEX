/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  async rewrites() {
    const apiTarget = process.env.INTERNAL_API_URL || 'http://api:8080';
    const cleanTarget = apiTarget.replace(/\/api\/?$/, '');
    return [
      {
        source: '/api/:path*',
        destination: `${cleanTarget}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
