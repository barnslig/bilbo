const autoprefixer = require('autoprefixer');
const cssnano = require('cssnano');
const mqpacker = require('css-mqpacker');
const pxtorem = require('postcss-pxtorem');
const sass = require('@csstools/postcss-sass');

module.exports = {
  map: false,
  plugins: [
    sass(),
    pxtorem({
      propList: ['*'],
    }),
    autoprefixer(),
    mqpacker({
      sort: true,
    }),
    cssnano(),
  ],
};
