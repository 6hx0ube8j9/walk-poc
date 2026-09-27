package main

import (
	"io"
	"log"
	"os"

	"github.com/tailscale/walk"
	. "github.com/tailscale/walk/declarative"
)

func main() {
	logFile, err := os.OpenFile("walk-poc.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	var mw *walk.MainWindow
	var ni *walk.NotifyIcon
	var isExiting bool

	// tailscale/walk 必须使用 InitApp 初始化
	app, err := walk.InitApp()
	if err != nil {
		log.Fatalf("初始化 Walk App 失败: %v", err)
	}

	err = MainWindow{
		AssignTo: &mw,
		Title:    "Tailscale Walk 生命周期测试",
		MinSize:  Size{Width: 400, Height: 240},
		Size:     Size{Width: 460, Height: 280},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}, Spacing: 12},
		Children: []Widget{
			Label{Text: "测试窗口生命周期"},
		},
	}.Create()
	if err != nil {
		log.Fatalf("创建窗口失败: %v", err)
	}

	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		log.Printf("[Event] Closing 触发: isExiting=%v", isExiting)
		if !isExiting {
			*canceled = true
			mw.SetVisible(false)
			log.Println("[Event] 已将 *canceled 置为 true 并 SetVisible(false)")
		}
	})

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
			if !mw.Visible() {
				mw.Show()
			}
			mw.SetFocus()
		}
	})

	exitAction := walk.NewAction()
	exitAction.SetText("彻底退出")
	exitAction.Triggered().Attach(func() {
		isExiting = true
		mw.Close()
	})
	ni.ContextMenu().Actions().Add(exitAction)

	mw.Show()
	log.Println("[App] 进入 app.Run() 消息循环")
	// tailscale/walk 唯一的合法入口
	exitCode := app.Run()
	log.Printf("[App] app.Run() 退出，退出码: %d", exitCode)
}
