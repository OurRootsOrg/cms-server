const CopyWebpackPlugin = require("copy-webpack-plugin");

module.exports = {
  publicPath: process.env.NODE_ENV === "production" ? "" : "/",
  transpileDependencies: ["vuetify"],
  configureWebpack: {
    plugins: [
      new CopyWebpackPlugin({
        patterns: [
          { from: "src/assets/seadragon", to: "img/seadragon" },
          { from: "node_modules/openseadragon/build/openseadragon/images", to: "img/seadragon" }
        ]
      })
    ]
  }
};
