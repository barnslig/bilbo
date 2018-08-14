const path = require('path');
const webpack = require('webpack');

module.exports = {
  entry: path.resolve('src/app.js'),
  module: {
    rules: [
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader'],
      },
      {
        exclude: /node_modules/,
        test: /\.js$/,
        use: ['babel-loader'],
      },
    ],
  },
  output: {
    filename: 'app.js',
    path: path.resolve('../static'),
    publicPath: '/static/',
  },
};
