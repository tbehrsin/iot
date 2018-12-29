const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const NgrokWebpackPlugin = require('ngrok-webpack-plugin');

module.exports = {
  entry: ['./src/index.scss', './src/index.js'],

  resolve: {
    alias: {
      'config': path.resolve(__dirname, 'app.json')
    }
  },

  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'site.js',
    publicPath: '/'
  },

  devtool: 'source-map',

  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /(node_modules|bower_components)/,
        use: {
          loader: 'babel-loader'
        }
      },
      {
        test: /\.scss$/,
        use: [
          "style-loader",
          { loader: "css-loader", options: { modules: true, camelCase: true } },
          { loader: 'resolve-url-loader' },
          { loader: "sass-loader", options: { sourceMap: true, sourceMapContents: false } },
        ],
        exclude: [
          path.resolve(__dirname, 'src', 'index.scss')
        ]
      },
      {
        include: path.resolve(__dirname, 'src', 'index.scss'),
        use: [
          "style-loader",
          { loader: "css-loader", options: { modules: false } },
          { loader: 'resolve-url-loader' },
          { loader: "sass-loader", options: { sourceMap: true, sourceMapContents: false } },
        ]
      },
      {
        test: /\.(eot|svg|ttf|woff2?|png|jpg|gif)$/,
        use: [
          {
            loader: 'file-loader',
            options: {
              name: 'assets/[name]-[hash].[ext]',
              publicPath: '/'
            }
          }
        ]
      }
    ]
  },

  devServer: {
    port: 8000
  },

  plugins: [
    new HtmlWebpackPlugin({
      template: 'src/index.html.ejs'
    })
    // new NgrokWebpackPlugin({
    //   subdomain: 'iot',
    //   region: 'eu'
    // })
  ]
};
