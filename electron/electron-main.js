'use strict'

const devMod = (process.argv.indexOf("--dev") >= 0)

const { app, Menu, BrowserWindow, dialog } = require('electron');

var log = require('electron-log');

const path = require('path');

const childProcess = require('child_process');


// This adds refresh and devtools console keybindings
// Page can refresh with cmd+r, ctrl+r, F5
// Devtools can be toggled with cmd+alt+i, ctrl+shift+i, F12
require('electron-debug')({ enabled: true, showDevTools: false });
require('electron-context-menu')({});


global.eval = function () { throw new Error('bad!!'); }

let walletPort = 8640;
let mePort = 8641;

let defaultURL = 'http://127.0.0.1:' + mePort + '/start.html';
global.sharedObject = {
  walletPort:walletPort,
  mePort: mePort
}
// let defaultURL;
// let filemanageURL;
// let port=7787;
// var portfinder = require('portfinder');
// portfinder.basePort=7788;
// portfinder.getPort(function (err, p) {
//   port = p;
//   defaultURL = 'http://127.0.0.1:'+port+'/';
//   filemanageURL = 'http://127.0.0.1:'+port+'/disk.html';
// });

let currentURL;


// Force everything localhost, in case of a leak
app.commandLine.appendSwitch('host-rules', 'MAP * 127.0.0.1, EXCLUDE *.store.samos.io, *.samos.io');
app.commandLine.appendSwitch('ssl-version-fallback-min', 'tls1.2');
app.commandLine.appendSwitch('--no-proxy-server');
app.setAsDefaultProtocolClient('samos');


var env = {}
env.PATH = [
  '$PATH'
  // Add third-party binaries paths here
].join(':')

// 3rd party libraries
env.DYLD_LIBRARY_PATH = [
  '$DYLD_LIBRARY_PATH',
  '/Applications/Samos-me.app/Contents/Resources/app/lib/',
  // Add more third-party lib paths here
].join(':')

process.env.DYLD_LIBRARY_PATH = env.DYLD_LIBRARY_PATH
//for macos

// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the JavaScript object is garbage collected.
let win;

 // Resolve binary location
var appPath = app.getPath('exe');
var exe;
if (!devMod) {
  exe = (() => {
    switch (process.platform) {
      case 'darwin':
        return path.join(appPath, '../../Resources/app/');
      case 'win32':
        // Use only the relative path on windows due to short path length
        // limits
        return './resources/app/';
      case 'linux':
        return path.join(path.dirname(appPath), './resources/app/');
      default:
        return './resources/app/';
    }
  })()
}
 
var nebula = null;
var wallet = null;
var nebulaStarted = false;
var walletStarted = false;


function startNebula() {
  console.log('Starting nebula from electron');
  console.log("PATH:"+process.env.DYLD_LIBRARY_PATH)

  if (nebula) {
    console.log('nebula already running');
    // if(wallet){
    //   app.emit('all-ready');
    // }
    return
  }

  var reset = () => {
    nebula = null;
  }

  // Resolve nebula-client binary location
  if (devMod) {
    exe = (() => {
      switch (process.platform) {
        case 'darwin':
          return path.join(path.dirname(appPath), '../../../../../../../client/nebula-client');
        case 'win32':
          // Use only the relative path on windows due to short path length
          // limits
          return '../client/nebula-client.exe';
        case 'linux':
          return path.join(path.dirname(appPath), '../../../../client/nebula-client');
        default:
          return './resources/app/nebula-client';
      }
    })()
  }

  var args = [
    '--launch-browser=false',
    '--webdir=' + path.dirname(exe) + '/web/build',
    '--server=127.0.0.1:' + mePort,
    '--collect=collector.store.samos.io:6688',
    '--tracker=tracker.store.samos.io:6677'
  ]
  nebula = childProcess.spawn(exe, args);

  nebula.on('error', (e) => {
    dialog.showErrorBox('Failed to start nebula', e.toString());
    app.quit();
  });

  nebula.stdout.on('data', (data) => {
    // console.log(data.toString());
    // Scan for the web URL string
    if (currentURL) {
      return
    }
    const marker = 'HTTP server listening on ';
    var i = data.indexOf(marker);
    if (i === -1) {
      return
    }
    nebulaStarted=true;
    if(walletStarted){
      currentURL = defaultURL;
      app.emit('all-ready', { url: currentURL });
    }
    console.log(data.toString());
  });

  nebula.stderr.on('data', (data) => {
    console.log(data.toString());
  });

  nebula.on('close', (code) => {
    console.log('nebula closed');
    reset();
  });

  nebula.on('exit', (code) => {
    console.log('nebula exited');
    reset();
  });
}


function startWallet() {
  console.log('Starting wallet from electron');
  if (wallet) {
    console.log('wallet already running');
    // if(nebula){
    //   app.emit('all-ready');
    // }
    return
  }
  var reset = () => {
    wallet = null;
  }
  if (devMod) {
    exe = (() => {
      switch (process.platform) {
        case 'darwin':
          return path.join(path.dirname(appPath), '../../../../../../../../samos/');
        case 'win32':
          // Use only the relative path on windows due to short path length
          // limits
          return '../../samos/';
        case 'linux':
          return path.join(path.dirname(appPath), '../../../../../samos/');
        default:
          return './resources/app/';
      }
    })()
  } 
  var args = [
    '-launch-browser=false',
    '-gui-dir=' + exe+"src/gui/static/",
    '-color-log=false', // must be disabled for web interface detection
    '-logtofile=true',
    '-download-peerlist=true',
    '-enable-seed-api=true',
    '-enable-wallet-api=true',
    '-rpc-interface=false',
    "-disable-csrf=false"
    // will break
    // broken (automatically generated certs do not work):
    // '-web-interface-https=true',
  ]
  wallet = childProcess.spawn(exe+"wallet", args);

  wallet.on('error', (e) => {
    console.log( e.toString());
    dialog.showErrorBox('Failed to start wallet', e.toString());
    app.quit();
  });

  wallet.stdout.on('data', function (data) {
    if (currentURL) {
      return
    }
    const marker = 'Starting web interface on ';
    var i = data.indexOf(marker);
    if (i === -1) {
      return
    }
    walletStarted=true;
    if(nebulaStarted){
      currentURL = defaultURL;
      app.emit('all-ready', { url: currentURL });
    }
    // console.log(data.toString());
  });

  wallet.stderr.on('data', (data) => {
    console.log(data.toString());
  });

  wallet.on('close', (code) => {
    console.log('wallet closed');
    reset();
  });

  wallet.on('exit', (code) => {
    console.log('wallet exited');
    reset();
  });
}


function createWindow(url) {
  if (!url) {
    url = defaultURL;
  }

  // To fix appImage doesn't show icon in dock issue.
  var iconPath = (() => {
    switch (process.platform) {
      case 'linux':
        if (!devMod) {
          return path.join(path.dirname(appPath), './resources/icon512x512.png');
        }else{
          return path.join(path.dirname(appPath), '../../../../assets/icon512x512.png');
        }
    }
  })()

  // Create the browser window.
  win = new BrowserWindow({
    width: 1200,
    height: 900,
    title: 'Samos ME',
    icon: iconPath,
    nodeIntegration: false,
    webPreferences: {
      webgl: false,
      webaudio: false,
    },
  });

  // patch out eval
  win.eval = global.eval;

  const ses = win.webContents.session
  ses.clearCache(function () {
    console.log('Cleared cache.');
  });

  ses.clearStorageData([], function () {
    console.log('Cleared the stored cached data');
  });

  win.loadURL(url);

  // Open the DevTools.
  // win.webContents.openDevTools();

  // Emitted when the window is closed.
  win.on('closed', () => {
    // Dereference the window object, usually you would store windows
    // in an array if your app supports multi windows, this is the time
    // when you should delete the corresponding element.
    win = null;
  });

  win.webContents.on('will-navigate', function (e, url) {
    e.preventDefault();
    require('electron').shell.openExternal(url);
  });

  // create application's main menu
  var template = [{
    label: "Samos",
    submenu: [
      { label: "Quit", accelerator: "Command+Q", click: function () { app.quit(); } }
    ]
  }, {
    label: "Edit",
    submenu: [
      { label: "Undo", accelerator: "CmdOrCtrl+Z", selector: "undo:" },
      { label: "Redo", accelerator: "Shift+CmdOrCtrl+Z", selector: "redo:" },
      { type: "separator" },
      { label: "Cut", accelerator: "CmdOrCtrl+X", selector: "cut:" },
      { label: "Copy", accelerator: "CmdOrCtrl+C", selector: "copy:" },
      { label: "Paste", accelerator: "CmdOrCtrl+V", selector: "paste:" },
      { label: "Select All", accelerator: "CmdOrCtrl+A", selector: "selectAll:" }
    ]
  }];

  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
}

// Enforce single instance
const alreadyRunning = app.makeSingleInstance((commandLine, workingDirectory) => {
  // Someone tried to run a second instance, we should focus our window.
  if (win) {
    if (win.isMinimized()) {
      win.restore();
    }
    win.focus();
  } else {
    createWindow(currentURL || defaultURL);
  }
});

if (alreadyRunning) {
  app.quit();
  return;
}

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
// Some APIs can only be used after this event occurs.
app.on('ready', function(){
  startWallet();
  startNebula();
});

app.on('all-ready', (e) => {
  createWindow(e.url);
});

// Quit when all windows are closed.
app.on('window-all-closed', () => {
  // On OS X it is common for applications and their menu bar
  // to stay active until the user quits explicitly with Cmd + Q
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  // On OS X it's common to re-create a window in the app when the
  // dock icon is clicked and there are no other windows open.
  if (win === null) {
    createWindow();
  }
});

app.on('will-quit', () => {
  if (nebula) {
    nebula.kill('SIGINT');
  }
  if (wallet) {
    wallet.kill('SIGINT');
  }
});
const { ipcMain } = require('electron')

ipcMain.on('close', e=> win.close());

const {shell} = require('electron')
ipcMain.on('explorer', (event,code) => {
  shell.openExternal("http://explorer.samos.io/app/address/"+code+"/1")
}); 