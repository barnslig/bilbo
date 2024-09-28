import pxtorem from "postcss-pxtorem";
import sortMediaQueries from "postcss-sort-media-queries";

export default {
  plugins: [
    pxtorem({
      propList: ["*"],
      selectorBlackList: ["html"],
    }),
    sortMediaQueries(),
  ],
};
