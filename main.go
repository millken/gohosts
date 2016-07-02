package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/MakeNowJust/hotkey"
	"github.com/cxfksword/gohosts/dist"
	"github.com/cxfksword/win"
	"github.com/oskca/sciter"
	"github.com/oskca/sciter/window"
)

var debug *bool = flag.Bool("vv", false, "enable debug")
var loadDist *bool = flag.Bool("dist", false, "load dist layout")
var disableHotKey *bool = flag.Bool("disableHotKey", false, "diable hot key")
var windowTitle string = "GoHosts v0.1"

func main() {
	flag.Parse()

	// limit only one instance running
	runingHwnd := win.FindWindow(nil, syscall.StringToUTF16Ptr(windowTitle))
	if runingHwnd != 0 {
		// has another gohosts process running
		win.ShowWindow(runingHwnd, win.SW_RESTORE)
		win.SetForegroundWindow(runingHwnd)
		return
	}

	screenWidth := int(win.GetSystemMetrics(win.SM_CXSCREEN))
	screenHeight := int(win.GetSystemMetrics(win.SM_CYSCREEN))
	width := 900
	height := 560
	x := int(screenWidth/2 - width/2)
	y := int(screenHeight/2 - height/2)
	// left, Top, Right, Bottom
	rect := sciter.NewRect(y, x, width, height)
	createFlags := sciter.SW_TITLEBAR | sciter.SW_RESIZEABLE | sciter.SW_CONTROLS | sciter.SW_MAIN
	if *debug {
		log.Println("[DEBUG MODE]")
		createFlags = createFlags | sciter.SW_ENABLE_DEBUG
	}
	w, err := window.New(createFlags, rect)
	if err != nil {
		log.Fatal(err)
	}

	w.DefineFunction("deleteFile", func(args ...*sciter.Value) *sciter.Value {
		path := args[0].String()
		err := os.Remove(path)
		if err != nil {
			log.Println(err)
			return sciter.NewValue(false)
		} else {
			return sciter.NewValue(true)
		}
	})
	w.DefineFunction("fileExists", func(args ...*sciter.Value) *sciter.Value {
		path := args[0].String()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return sciter.NewValue(false)
		} else {
			return sciter.NewValue(true)
		}
	})
	w.DefineFunction("renameFile", func(args ...*sciter.Value) *sciter.Value {
		oldPath := args[0].String()
		newPath := args[1].String()
		err := os.Rename(oldPath, newPath)
		if err != nil {
			log.Println(err)
			return sciter.NewValue(false)
		} else {
			return sciter.NewValue(true)
		}
	})
	w.DefineFunction("mkdir", func(args ...*sciter.Value) *sciter.Value {
		path := args[0].String()
		os.MkdirAll(path, os.ModePerm)
		return sciter.NewValue(true)
	})
	w.DefineFunction("clearDNSCache", func(args ...*sciter.Value) *sciter.Value {
		if runtime.GOOS == "windows" {
			cmd := exec.Command("ipconfig", "/flushdns")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			err := cmd.Start()
			if err != nil {
				log.Println(err)
				return sciter.NewValue(false)
			} else {
				return sciter.NewValue(true)
			}
		}

		return sciter.NewValue(false)
	})
	w.DefineFunction("clipboardText", func(args ...*sciter.Value) *sciter.Value {
		var hwnd win.HWND
		if !win.OpenClipboard(hwnd) {
			log.Println("OpenClipboard error")
			return sciter.NewValue("")
		}
		defer win.CloseClipboard()

		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_UNICODETEXT))
		if hMem == 0 {
			log.Println("GetClipboardData error")
			return sciter.NewValue("")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			log.Println("GlobalLock() error")
			return sciter.NewValue("")
		}
		defer win.GlobalUnlock(hMem)

		text := win.UTF16PtrToString((*uint16)(p))

		return sciter.NewValue(text)
	})

	rootPath, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(rootPath, "res/app.htm")); os.IsNotExist(err) || *loadDist {
		log.Println("[LOAD DIST LAYOUT]")
		w.LoadHtml(dist.DeployBinData, "")
		// w.LoadFile("dist/app.htm")
	} else {
		w.LoadFile("res/app.htm")
	}
	w.SetTitle(windowTitle)

	if runtime.GOOS == "windows" {
		// set icon
		hwnd := win.HWND(unsafe.Pointer(w.GetHwnd()))
		// absFilePath, _ := filepath.Abs("app.ico")
		// hIcon := win.HICON(win.LoadImage(
		// 	0,
		// 	syscall.StringToUTF16Ptr(absFilePath),
		// 	win.IMAGE_ICON,
		// 	0,
		// 	0,
		// 	win.LR_DEFAULTSIZE|win.LR_LOADFROMFILE))
		hIcon := NewIconFromResource("GLFW_ICON")
		if hIcon != 0 {
			win.SendMessage(hwnd, win.WM_SETICON, 1, uintptr(unsafe.Pointer(hIcon)))
			win.SendMessage(hwnd, win.WM_SETICON, 0, uintptr(unsafe.Pointer(hIcon)))
		}
	}

	registerHotKey(w)
	w.Show()
	w.Run()
}

// NewIconFromResource returns a new Icon, using the specified icon resource.
func NewIconFromResource(resName string) (hIcon win.HICON) {
	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		hIcon = 0
		log.Println("GetModuleHandle() error")
		return
	}
	if hIcon = win.LoadIcon(hInst, syscall.StringToUTF16Ptr(resName)); hIcon == 0 {
		log.Println("LoadIcon() error")
	}

	return
}

func NewIconFromResourceId(id uintptr) (hIcon win.HICON) {
	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		hIcon = 0
		log.Println("GetModuleHandle() error")
		return
	}

	if hIcon = win.LoadIcon(hInst, win.MAKEINTRESOURCE(id)); hIcon == 0 {
		log.Println("LoadIcon() error")
	}

	return
}

// ctrl+ctrl to active windows
func registerHotKey(w *window.Window) {
	if *disableHotKey {
		return
	}

	var prevHotKeyPressTime time.Time = time.Date(2000, time.January, 1, 1, 0, 0, 0, time.Local)

	hkey := hotkey.New()
	hkey.Register(hotkey.Ctrl, 0, func() {
		if time.Now().Before(prevHotKeyPressTime.Add(400 * time.Millisecond)) {
			hwnd := win.HWND(unsafe.Pointer(w.GetHwnd()))
			curHwnd := win.GetForegroundWindow()
			if curHwnd != hwnd {
				win.ShowWindow(hwnd, win.SW_RESTORE)
				win.SetForegroundWindow(hwnd)
			} else {
				win.ShowWindow(hwnd, win.SW_MINIMIZE|win.SW_HIDE)
			}

			prevHotKeyPressTime = time.Date(2000, time.January, 1, 1, 0, 0, 0, time.Local)
			return
		}

		prevHotKeyPressTime = time.Now()
		return
	})
}
