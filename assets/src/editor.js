import Axios from 'axios';

import CodeMirror from 'codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/gfm/gfm';
import 'codemirror/theme/monokai.css';

Axios.defaults.headers.post['X-CSRF-Token'] =
  document.querySelector('meta[name="csrf-token"]').getAttribute('content');

class Editor {
  constructor($elem) {
    this.ui = {
      editor: $elem.querySelector('[data-editor-edit]'),
      form: $elem.querySelector('[data-editor-form]'),
      loader: $elem.querySelector('[data-editor-loader]'),
      message: $elem.querySelector('[data-editor-message]'),
      preview: $elem.querySelector('[data-editor-preview]'),
      self: $elem,
    };

    this.isDirty = false;
    this.isUpdating = false;
    this.lastContent = '';

    this.cm = new CodeMirror.fromTextArea(this.ui.editor, {
      autofocus: true,
      lineNumbers: true,
      lineWrapping: true,
      mode: 'gfm',
      theme: 'monokai',
    });

    this.cm.on('changes', () => {
      this.isDirty = true;
    })

    this.ui.form.addEventListener('submit', this.onSubmit.bind(this));

    window.addEventListener('beforeunload', this.onUnload.bind(this));

    setInterval(this.onUpdatePreview.bind(this), 1000);
    this.onUpdatePreview();
  }

  set isLoaderVisible(isVisible) {
    const className = 'm-editor-header__loader--visible';
    this.ui.loader.classList[isVisible ? 'add' : 'remove'](className);
  }

  set isLoading(isLoading) {
    const showAfter = 500;
    clearTimeout(this.loaderTimeout);
    if (isLoading) {
      this.loaderTimeout = setTimeout(() => this.isLoaderVisible = true, showAfter);
    } else {
      this.isLoaderVisible = false;
    }
  }

  onUnload(ev) {
    if (this.isDirty) {
      ev.preventDefault();
      const msg = 'Leave page and destroy changes?';
      event.returnValue = msg;
      return msg;
    }
  }

  onSubmit(ev) {
    ev.preventDefault();

    this.isLoading = true;

    Axios
      .post(this.ui.form.target, {
        data: this.cm.getValue(),
        message: this.ui.message.value,
      })
      .then(response => {
        this.isDirty = false;
        window.location = this.ui.self.dataset.linkpath;
      });
  }

  onUpdatePreview() {
    const value = this.cm.getValue();
    if (!this.isUpdating && value !== this.lastContent) {
      this.isLoading = true;
      this.isUpdating = true;
      this.lastContent = value;

      const url = '/edit/_preview';

      Axios
        .post(url, {
          data: this.lastContent,
          filepath: this.ui.self.dataset.filepath,
        })
        .then(response => {
          this.isLoading = false;
          this.isUpdating = false;
          this.ui.preview.innerHTML = response.data;
        });
    } else {
      this.isLoading = false;
    }
  }
}

export default Editor;
