const { GenerateSW } = require('workbox-webpack-plugin')

module.exports = {
  lintOnSave: false,
  productionSourceMap: false,
  publicPath: process.env.NODE_ENV === 'development' ? '/vuejs-pwa/' : '',
  configureWebpack: {
    plugins: [new GenerateSW()]
  },
  devServer: {
    proxy: {
      '/ws/': {
        target: 'http://localhost:8080',
        ws: true,
        changeOrigin: false,
        onProxyReq: function (request) {
          request.setHeader('origin', 'http://localhost:8080')
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
