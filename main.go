package main

import (
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/tailscale/walk"
	. "github.com/tailscale/walk/declarative"
)

func main() {
	logFile, err := os.OpenFile("walk-poc.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[FATAL PANIC] 异常崩溃: %v\n%s", r, debug.Stack())
		}
		log.Println("[App] 进程完全退出")
	}()

	var mw *walk.MainWindow
	var ni *walk.NotifyIcon
	var isExiting bool

	// 1. 创建主窗口 (2022 年底版本仍是原生 MainWindow 体系，无需 InitApp)
	err = MainWindow{
		AssignTo: &mw,
		Title:    "Tailscale Walk df8c773 生命周期测试",
		MinSize:  Size{Width: 400, Height: 240},
		Size:     Size{Width: 460, Height: 280},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}, Spacing: 12},
		Children: []Widget{
			Label{
				Text: "测试验证：\n点击右上角 [X]，验证 2022 年底版本的原生 Closing 拦截表现。",
			},
		},
	}.Create()
	if err != nil {
		log.Fatalf("创建窗口失败: %v", err)
	}

	// 2. 原生 Closing 拦截
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		log.Printf("[Event] Closing 触发: isExiting=%v", isExiting)
		if !isExiting {
			*canceled = true
			mw.SetVisible(false)
			log.Println("[Event] 成功拦截 [X]，窗口已隐藏")
		} else {
			log.Println("[Event] 放行退出")
		}
	})

	// 3. 创建托盘 (df8c773 仍需要传入宿主窗口 mw)
	ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatalf("创建托盘失败: %v", err)
	}
	defer ni.Dispose()

	ni.SetIcon(walk.IconApplication())
	ni.SetToolTip("Tailscale df8c773 POC")
	ni.SetVisible(true)

	ni.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			log.Println("[Tray] 左键点击托盘，恢复窗口")
			if !mw.Visible() {
				mw.Show()
			}
			mw.SetFocus()
		}
	})

	exitAction := walk.NewAction()
	exitAction.SetText("彻底退出")
	exitAction.Triggered().Attach(func() {
		log.Println("[Tray] 用户点击退出")
		isExiting = true
		mw.Close()
	})
	ni.ContextMenu().Actions().Add(exitAction)

	// 4. 显示窗口并使用 mw.Run() 启动消息循环
	mw.Show()
	log.Println("[App] 进入 mw.Run() 消息循环")
	exitCode := mw.Run()
	log.Printf("[App] mw.Run() 循环退出，退出码: %d", exitCode)
}
