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

	// 1. 创建主窗口 (不调用 walk.InitApp)
	err = MainWindow{
		AssignTo: &mw,
		Title:    "Tailscale Walk 生命周期测试 (mw.Run 模式)",
		MinSize:  Size{Width: 400, Height: 240},
		Size:     Size{Width: 460, Height: 280},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}, Spacing: 12},
		Children: []Widget{
			Label{
				Text: "测试验证：\n点击右上角 [X]，验证 mw.Run() 模式下是否还会被强制杀死。",
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
			log.Println("[Event] 窗口已隐藏，拦截成功")
		} else {
			log.Println("[Event] 放行退出")
		}
	})

	// 监控窗口是否真的被 Win32 底层销毁
	mw.Disposing().Attach(func() {
		log.Println("[Event] 警告: MainWindow 句柄正在被销毁 (Disposed)")
	})

	// 3. 创建托盘图标 (tailscale/walk 原生支持无参数调用)
	ni, err = walk.NewNotifyIcon()
	if err != nil {
		log.Fatalf("创建托盘失败: %v", err)
	}
	defer ni.Dispose()

	ni.SetIcon(walk.IconApplication())
	ni.SetToolTip("Tailscale POC")
	ni.SetVisible(true)

	ni.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			log.Println("[Tray] 左键点击，恢复窗口")
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

	// 4. 显示窗口并使用 mw.Run() 进入消息循环
	mw.Show()
	log.Println("[App] 进入 mw.Run() 消息循环")
	exitCode := mw.Run()
	log.Printf("[App] mw.Run() 循环退出，退出码: %d", exitCode)
}
