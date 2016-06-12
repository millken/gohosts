//go:generate rsrc -arch=amd64 -ico=app.ico
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"github.com/oskca/sciter"
	"github.com/oskca/sciter/window"
)

var debug *bool = flag.Bool("vv", false, "enable debug")

func main() {
	flag.Parse()

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
	w.LoadFile("app.htm")
	w.SetTitle("GoHosts v0.1")

	if runtime.GOOS == "windows" {
		// set icon
		hwnd := win.HWND(unsafe.Pointer(w.GetHwnd()))
		absFilePath, _ := filepath.Abs("app.ico")
		hIcon := win.HICON(win.LoadImage(
			0,
			syscall.StringToUTF16Ptr(absFilePath),
			win.IMAGE_ICON,
			0,
			0,
			win.LR_DEFAULTSIZE|win.LR_LOADFROMFILE))
		if hIcon != 0 {
			win.SendMessage(hwnd, win.WM_SETICON, 1, uintptr(unsafe.Pointer(hIcon)))
			win.SendMessage(hwnd, win.WM_SETICON, 0, uintptr(unsafe.Pointer(hIcon)))
		}
	}

	w.Show()
	w.Run()
}
