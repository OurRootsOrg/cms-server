const CopyWebpackPlugin = require("copy-webpack-plugin");

module.exports = {
  transpileDependencies: ["vuetify"],
  configureWebpack: {
    plugins: [
      new CopyWebpackPlugin({
        patterns: [
          { from: "node_modules/oidc-client/dist/oidc-client.min.js", to: "js" },
          { from: "src/assets/images", to: "img/images" },
          { from: "src/assets/seadragon", to: "img/seadragon" },
          { from: "node_modules/openseadragon/build/openseadragon/images", to: "img/seadragon" }
        ]
      })
    ]
  }
};
