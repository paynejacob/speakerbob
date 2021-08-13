const { GenerateSW } = require('workbox-webpack-plugin')
const PreloadWebpackPlugin = require('@vue/preload-webpack-plugin')

module.exports = {
  lintOnSave: false,
  productionSourceMap: false,
  configureWebpack: {
    plugins: [
      new GenerateSW(),
      new PreloadWebpackPlugin({
        rel: 'preload',
        as(entry) {
          if (/\.css$/.test(entry)) return 'style';
          if (/\.woff$/.test(entry)) return 'font';
          if (/\.woff2$/.test(entry)) return 'font';
          if (/\.png$/.test(entry)) return 'image';
          return 'script';
        }
      })]
  },
  devServer: {
    proxy: {
      '/ws/': {
        target: 'http://127.0.0.1:8080',
        ws: true,
        changeOrigin: false,
        onProxyReq: function (request) {
          request.setHeader('origin', 'http://127.0.0.1:8080')
        }
      },
      '/sound/': {
        target: 'http://localhost:8080'
      },
      '/sound/sound/': {
        target: 'http://localhost:8080'
      },
      '/play/': {
        target: 'http://localhost:8080'
      }
    }
  }
}
