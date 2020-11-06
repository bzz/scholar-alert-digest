const path = require("path")
const webpack = require("webpack")
const HtmlWebpackPlugin = require("html-webpack-plugin")

module.exports = env => ({
  mode: env.development ? "development" : "production",
  entry: [
    "regenerator-runtime/runtime",
    "./src/index.js",
  ],
  output: {
    filename: "index.js",
    path: path.resolve(__dirname, "dist"),
  },
  resolve: {
    modules: [path.resolve(__dirname, "src"), "node_modules"],
  },
  devtool: "inline-source-map",
  devServer: {
    contentBase: path.join(__dirname, "dist"),
    port: 9000,
  },
  plugins: [
    new HtmlWebpackPlugin({
      favicon: "./favicon.png",
    }),
    new webpack.DefinePlugin({
      __DEV__: env.development,
      __PROD_BASE__: JSON.stringify(env.baseUrl),
    }),
  ],
  module: {
    rules: [
      {
        test: /\.m?js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader",
          options: {
            presets: ["@babel/preset-env", "@babel/preset-react"],
          },
        },
      },
      {
        test: /\.css$/i,
        use: ["style-loader", "css-loader"],
      },
    ],
  },
})
