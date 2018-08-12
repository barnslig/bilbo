const autoprefixer = require('autoprefixer');
const cssnano = require('cssnano');
const mqpacker = require('css-mqpacker');
const sass = require('@csstools/postcss-sass');

module.exports = {
  map: false,
  plugins: [
    sass(),
    mqpacker({
      sort: true,
    }),
    autoprefixer(),
    cssnano(),
  ],
};
