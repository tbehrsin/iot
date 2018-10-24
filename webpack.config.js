const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = [
  {
    entry: './src/electron/index.js',
    output: {
      filename: 'electron.js',
      path: path.resolve(__dirname, 'dist')
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          exclude: /node_modules/,
          use: 'babel-loader'
        }
      ],
    },
    target: 'electron-main'
  },
  {
    entry: './renderer.js',
    context: path.resolve(__dirname, 'src', 'electron'),
    output: {
      filename: 'renderer.js',
      publicPath: '/',
      path: path.resolve(__dirname, 'dist')
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          exclude: /node_modules/,
          use: 'babel-loader'
        },
        {
          test: /\.scss$/,
          use: [
              "style-loader",
              "css-loader",
              "sass-loader"
          ]
        },
        {
          test: /\.(ttf|png|jpe?g)$/,
          use: {
            loader: 'file-loader',
            options: {
              name: '[path][name].[ext]',
            }
          }
        }
      ],
    },
    target: 'electron-renderer',
    plugins: [
      new HtmlWebpackPlugin({
        template: './index.html.ejs',
        minify: true
      })
    ]
  }
];
