const { GenerateSW } = require('workbox-webpack-plugin')
const PreloadWebpackPlugin = require('@vue/preload-webpack-plugin')

webpackPlugins = []

if (process.env.NODE_ENV === "production") {
  webpackPlugins = [
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
}

module.exports = {
  lintOnSave: false,
  productionSourceMap: false,
  configureWebpack: {
    plugins: webpackPlugins
  },
  devServer: {
    disableHostCheck: true,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        ws: true,
        changeOrigin: false,
        onProxyReq: function (request) {
          request.setHeader('origin', 'http://127.0.0.1:8080')
        }
      },
      '/auth': {
        target: 'http://127.0.0.1:8080'
      },
    }
  }
}
