package main

import (
	"io"
	"log"
	"os"

	// 核心改变：代码层面全部使用 lxn 原版路径
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	logFile, _ := os.OpenFile("walk-poc.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	var mw *walk.MainWindow
	var ni *walk.NotifyIcon
	var isExiting bool

	// 因为有了 go.mod 的 replace，这里的 InitApp 实际上调用的是 tailscale 的代码
	app, _ := walk.InitApp()

	MainWindow{
		AssignTo: &mw,
		Title:    "Tailscale Walk 终极实力局 (Replace模式)",
		Size:     Size{Width: 460, Height: 280},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "无需任何 SetWindowLongPtr，直接拦截 [X] 保持存活。"},
		},
	}.Create()

	// 1. 完全原版的拦截方式
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if !isExiting {
			*canceled = true
			mw.SetVisible(false)
			log.Println("[Event] [X] 拦截成功，窗口句柄存活，已安全隐藏")
		}
	})

	ni, _ = walk.NewNotifyIcon()
	defer ni.Dispose()
	ni.SetIcon(walk.IconApplication())
	ni.SetToolTip("Tray App")
	ni.SetVisible(true)

	ni.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			if !mw.Visible() {
				mw.Show()
			}
			mw.SetFocus()
		}
	})

	exitAction := walk.NewAction()
	exitAction.SetText("彻底退出")
	exitAction.Triggered().Attach(func() {
		log.Println("[Tray] 用户主动退出")
		isExiting = true
		app.Exit(0) 
	})
	ni.ContextMenu().Actions().Add(exitAction)

	mw.Show()

	// 2. 拿出实力：Win32 消息循环原地复活
	for !isExiting {
		log.Println("[App] (重)启动 app.Run() 消息循环...")
		app.Run()
		
		if !isExiting {
			log.Println("[App] 捕获到 Tailscale 的假死退出，窗口毫发无损，消息循环原地复活！")
		}
	}
	
	log.Println("[App] 进程真正彻底退出")
}
