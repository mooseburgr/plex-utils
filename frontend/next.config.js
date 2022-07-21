/** @type {import('next').NextConfig} */

const nextConfig = {
  reactStrictMode: true,
  images: {
    domains: ['lh3.googleusercontent.com'],
  },
  assetPrefix: './',
  async rewrites() {
    return [
      {
        source: '/cr',
        destination: 'https://cloudrun-srv-d3zjrv6fua-uc.a.run.app',
      },
    ];
  },
  async redirects() {
    return [
      {
        source: '/pw',
        destination: 'https://www.polywork.com/mooseburger',
        permanent: false,
      },
    ];
  }
};

module.exports = nextConfig
