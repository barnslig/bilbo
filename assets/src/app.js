const $editor = document.querySelector('[data-editor]');
if ($editor) {
  import(/* webpackChunkName: "editor" */ './editor')
    .then(({ default: Editor }) => new Editor($editor));
}
