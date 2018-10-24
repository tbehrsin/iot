import path from 'path';
import { app, BrowserWindow } from 'electron';

let win;

const createWindow = () => {
  win = new BrowserWindow({
    width: 800,
    height: 480,
    resizable: false,
    maximizable: false,
    kiosk: process.env.NODE_ENV === 'production'
  });
  win.loadURL('http://localhost:3000/');
};

app.on('ready', createWindow);
