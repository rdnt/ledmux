const path = require("path");

module.exports = {
  outputDir: path.resolve(__dirname, "../static/"),
  chainWebpack: config => {
    config.resolve.symlinks(false);
    config.module.rule('eslint').use('eslint-loader').options({
      fix: true,
    });
  },
  devServer: {
    host: "localhost"
  }
};
