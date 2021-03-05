module.exports = {
  lintOnSave: false,
  productionSourceMap: false,
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
      }
    }
  }
}
