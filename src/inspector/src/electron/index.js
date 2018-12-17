
const { app, BrowserWindow } = require('electron');

let win;

function createWindow () {
  win = new BrowserWindow({ width: 800, height: 600, title: 'z3js' })

  const host = 'localhost:9222';

  win.loadURL(`chrome-devtools://devtools/bundled/inspector.html?experiments=true&v8only=true&ws=${host}/`);
  win.webContents.on('did-finish-load', () => {
    win.webContents.executeJavaScript(`(() => {
      const tabbedPane = UI.inspectorView._tabbedPane;
      if (tabbedPane) {
        tabbedPane.closeTab('elements');
        tabbedPane.closeTab('security');
        tabbedPane.closeTab('timeline');
        tabbedPane.closeTab('network');
        tabbedPane.closeTab('audits2');
        tabbedPane.closeTab('resources');


        tabbedPane._leftToolbar._contentElement.remove();
        tabbedPane._rightToolbar._contentElement.remove();
      }
    })()`);
  });

  win.on('closed', () => {
    win = null;
  })
};

app.on('ready', createWindow);

app.on('window-all-closed', () => {
  app.quit();
});

app.on('activate', () => {
  if (win === null) {
    createWindow();
  }
});
